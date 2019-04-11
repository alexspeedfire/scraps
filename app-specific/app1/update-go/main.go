package main

import (
	"archive/zip"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"./config"
	"./files"

	ftpclient "github.com/jlaffaye/ftp"
)

const version string = "0.3" // Версия приложения

func main() {

	// Флаги для настройки по без необходимости править конфиг.
	custConf := flag.String("config", "config.yml", "Позволяет задать путь к конфигу.")
	custForce := flag.Bool("force", false, "Позволяет принудительно обновить версию без задания опции в конфиге.")
	custFirstRun := flag.Bool("firstrun", false, "Устанавливает режим первого запуска.")
	custUpdateOnly := flag.Bool("norun", false, "Запрещает запуск вектора после проверки обновления.")
	// TODO: Доделать фунционал dry-run
	//	custCheck := flag.Bool("dry-run", false, "Запускает приложение в режиме проверки, изменения не вносятся.")
	// Парсинг флагов
	flag.Parse()

	var localSize uint64     // размер локального файла
	localTime := time.Time{} // время модификации локального файла
	needUpdate := false      // требуется ли обновление

	log.Printf("Запущена проверка обновления app1, версия " + version)
	// Загружаем настройки из файла yml
	log.Println("Выполняется загрузка настроек.")
	log.Println("Настройки загружаются из файла", *custConf)
	conf := new(config.Configuration)
	err := conf.ReadConfig(*custConf)
	if err != nil {
		log.Printf("Ошибка при загрузке настроек: %s\n", err)
		os.Exit(1)
	}

	// Проверяем загруженные параметры
	if _, err := os.Stat(conf.Local.Path); os.IsNotExist(err) {
		log.Println("Путь к app1 не найден. Создаю.")
		os.MkdirAll(conf.Local.Path, 0755)
	}

	if _, err := os.Stat(conf.Local.Conf); os.IsNotExist(err) {
		log.Println("Путь к конфигам не найден. Создаю.")
		os.MkdirAll(conf.Local.Conf, 0755)
		log.Println("После установки требуется скопировать конфиги вручную.")
		conf.FirstRun = true
	}

	if _, err := os.Stat(conf.Local.Bin); os.IsNotExist(err) {
		log.Println("Путь к стартовыми файлами app1 не найден. Создаю.")
		os.MkdirAll(conf.Local.Bin, 0755)
		log.Println("После установки требуется скопировать стартовые файлы app1 вручную.")
		conf.FirstRun = true
	}

	if _, err := os.Stat(conf.Update.Log); os.IsNotExist(err) {
		log.Println("Путь к логам не найден. Создаю.")
		os.MkdirAll(conf.Update.Log, 0755)
	}

	if _, err := os.Stat(conf.Update.Archive); os.IsNotExist(err) {
		log.Println("Путь для резервной копии не найден. Создаю.")
		os.MkdirAll(conf.Update.Archive, 0755)
	}

	if conf.Update.Name == "" {
		log.Println("Не задан формат имени архива, используется значение по умолчанию.")
		conf.Update.Name = time.Now().Format("2006_01_02_15_04")
	} else {
		conf.Update.Name = time.Now().Format(conf.Update.Name)
	}

	if conf.System.Timeout == 0 {
		log.Println("Не задан таймаут связи с фтп сервером, используется значение по умолчанию.")
		conf.System.Timeout = 10
	}

	// Открываем (и закрываем) файл журнала.
	logFile, err := os.OpenFile(filepath.Join(conf.Update.Log, "log_"+conf.Update.Name+".log"), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	defer logFile.Close()
	if err != nil {
		// Ошибка открытия, ругаемся и просто пишем в консоль.
		log.Printf("Ошибка записи в лог файл: %s\n", err)
		log.Printf("Логи будут выводится только в консоль.")
	} else {
		// Быстро дампим часть строк в свежеоткрытый лог
		log.SetOutput(logFile)
		log.Printf("Запущена проверка обновления app1, версия " + version)
		log.Println("Настройки загружаются из файла", *custConf)
		// Мультиписалка, чтобы одновременно писать в лог и в консоль
		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
	}
	// Все ошибки чтения конфига отрабатываются в config.go.
	// TODO: Переделать, чтобы из конфига возвращались коды ошибок.
	log.Printf("Настройки успешно загружены, начинаю работу по обновлению.")

	// Переопределяем часть переменных флагами.
	if *custForce {
		// Принудительное обновление
		conf.System.Force = true
	}

	if *custFirstRun {
		// Первый запуск.
		// TODO: Переделать логику работы этой штуки.
		conf.FirstRun = true
	}

	// получаем имя файла сборки.
	tmp := strings.Split(conf.Remote.Path, "/")
	// Она должна быть в конце remote.path
	archname := tmp[len(tmp)-1]
	// TODO: Проверить - оно вообще имя файла или как.

	// Получаем время и размер локального файла
	localStat, err := os.Stat(conf.Local.Path + "/" + archname)
	if err != nil {
		if os.IsNotExist(err) {
			// Ошибка - файл не существует. Поставим размер в 0, время в текущее
			// и флаг принудительного обновления.
			log.Print("Локальный файл не существует, принудительное обновление.")
			localSize = 0
			localTime = time.Time{}
			needUpdate = true
		} else {
			// Какая-та другая ошибка. Ругаемся и сворачиваемся :(
			log.Fatalf("Ошибка при получении информации о файле (но файл существует): %s\n", err)
		}
	} else {
		// Если ошибок нет, записываем себе размер и дату модификации файла.
		// При этом размер конвертируем в uint64, чтобы потом не иметь мозг с сравнением размера с фтп
		localSize = uint64(localStat.Size())
		localTime = localStat.ModTime()
	}

	// Пытаемся подключиться к фтп серверу из конфига.
	conn, err := ftpclient.DialTimeout(conf.Remote.Host+":21", time.Duration(conf.System.Timeout)*time.Second)
	defer conn.Quit()
	if err != nil {
		// Подключиться не удалось, ругаемся и падаем.
		log.Fatalf("Ошибка подключения к ФТП серверу: %s.\n", err)
	}

	// Пытаемся залогиниться с известными учётными данными.
	err = conn.Login(conf.Remote.User, conf.Remote.Pass)
	if err != nil {
		// Войти не получилось, ругаемся и падаем.
		log.Fatalf("Ошибка входа на ФТП сервер: %s\n", err)
	}

	// Пытаемся перейти в папку с дистрибутивом.
	err = conn.ChangeDir(strings.TrimSuffix(conf.Remote.Path, archname))
	if err != nil {
		// Сменить путь не вышло, ругаемся и падаем.
		log.Fatalf("Ошибка при навигации по ФТП серверу: %s\n", err)
	}

	// Запрашиваем листинг с именем архива, должно вернуться 1.
	file, err := conn.List(archname)
	if err != nil {
		// Получить информацию о файле не вышло (но сюда мы не попадём, если файл не существует!)
		// Ругаемся и падаем.
		log.Fatalf("Ошибка при запросе информации о файле дистрибутива: %s\n", err)
	}

	// Если в ответ пришёл ноль - файла нет, ругаемся и падаем.
	if len(file) == 0 {
		log.Fatalf("Файл не найден в папке на фтп сервере.")
	}

	// Не уверен, что могло прилететь больше 1. Но на всякий случай ругаемся и падаем.
	if len(file) > 1 {
		log.Fatal("Получен неожиданный листинг (количество файлов больше одного).")
	}

	// Если всё раньше не свалились, записываем время и размер файла.
	// TODO: Добавить хеширование из версии 0.01.
	remoteTime := file[0].Time
	remoteSize := file[0].Size

	// Сравниваем время с сдвигом времени.
	// сдвиг - костыль для странного поведения сервера
	// в текущей конфигурации он возвращает правильное время
	// но неверную таймзону (GMT)
	if (remoteTime.Sub(localTime).Hours() > conf.System.Shift) && !needUpdate {
		log.Println("Файл на фтп новее, требуется обновление.")
		needUpdate = true
	}
	// Сравниваем размеры файлов.
	if (localSize != remoteSize) && !needUpdate {
		log.Println("Файлы разного размера, требуется обновление.")
		needUpdate = true
	}
	// Ну или просто принудительно взводим флаг обновления.
	if conf.System.Force {
		log.Println("В конфиге задана опция принудительного обновления.")
		needUpdate = true
	}

	// Если мы раньше поставили необходимость обновления - обновляемся.
	if needUpdate {
		log.Println("Запуск процедуры обновления")

		// Открываем соединение с сервером.
		ftpRert, err := conn.Retr(archname)
		defer ftpRert.Close()
		log.Println("Начинаю скачивать файл с фтп сервера " + conf.Remote.Host)
		if err != nil {
			// Ругаемся и падаем, если не получилось.
			log.Fatalf("Ошибка при получении файла: %s\n", err)
		}

		// Пытаемся переписать локальный файл.
		localFile, err := os.Create(filepath.Join(conf.Local.Path, archname))
		defer localFile.Close()
		if err != nil {
			// Увы, не получилось. Ругаемся и падаем
			log.Fatalf("Ошибка при работе с файловой системой: %s\n", err)
		}

		// Непосредственно качаем файл из ftpRert в localFile
		fileSize, err := io.Copy(localFile, ftpRert)
		if err != nil {
			// Увы, не получилось. Ругаемся и падаем.
			log.Fatalf("Ошибка при записи файл: %s\n", err)
		}

		// Сравниваем, сколько скопировалось с известным размером.
		if remoteSize != uint64(fileSize) {
			log.Fatalf("Размер скачанного файла не совпадает с размером на фтп.")
		}
		// Пишем сколько скопировали.
		log.Printf("Скопировано %d байт.\n", fileSize)

		// Если это не первый запуск
		if !conf.FirstRun {
			// Если требуется сжатие старых файлов, работаем над этим.
			if conf.Update.Compress {
				log.Println("Требуется сжатие архивной копии, подготавливаю файлы.")
				// Проверяем - есть ли вообще папка, чтобы не делать ненужную работу.
				if _, err := os.Stat(filepath.Join(conf.Local.Path, "app1")); os.IsNotExist(err) {
					log.Print("Не найдена установленная версия, пропускаю сжатие.")
				} else {
					// Создаём файл с архивом.
					outfile, err := os.Create(filepath.Join(conf.Update.Archive, "app1_"+conf.Update.Name+".zip"))
					defer outfile.Close()
					if err != nil {
						// Если не получилось - ругаемся и падаем.
						log.Fatalf("Ошибка при создании архивного файла: %s\n", err)
					}
					// Создаем райтер и пакуем всё по пути path + app1 в наш архив.
					w := zip.NewWriter(outfile)
					err = files.ZipFiles(w, conf.Local.Path+"/app1/", "")
					if err != nil {
						// Если архивация вернула ошибку, руаемся, скидываем флаг и просто копируем.
						log.Printf("Ошибка при архивации файлов: %s\n", err)
						conf.Update.Compress = false

						// Заодно удалим созданный файл.
						os.Remove(filepath.Join(conf.Update.Archive, "app1_"+conf.Update.Name+".zip"))

					} else {
						// Если получилось запаковать.
						// Удаляем рекурсивно папку с старой версией.
						err = os.RemoveAll(filepath.Join(conf.Local.Path, "app1"))
						if err != nil {
							// Если не вышло - паникуем и падаем.
							log.Fatalf("Ошибка при удалении установленной версии: %s\n", err)
						}
					}

					// Пытаемся закрыть райтер.
					err = w.Close()
					if err != nil {
						// Если не вышло - паникуем и падаем.
						log.Fatalf("Ошибка при закрытии файла архива: %s\n", err)
					}

				}
			}

			// Если сжимать не требуется  - переносим в нужную папку.
			if !conf.Update.Compress {
				log.Println("Сжатие архивной копии не требуется, подготавливаю перенос в архив.")
				// Проверяем наличие папки с утсановленной версией.
				if _, err := os.Stat(filepath.Join(conf.Local.Path, "app1")); os.IsNotExist(err) {
					// Если не находим - просто пишем и всё.
					log.Print("Не найдена установленная версия, пропускаю копирование")
				} else {
					// Если находим - пытаемся переименовать (и переместить, да)
					err = os.Rename(filepath.Join(conf.Local.Path, "app1"), filepath.Join(conf.Update.Archive, "app1_"+conf.Update.Name))
					if err != nil {
						// Если не вышло - падаем и паникуем.
						log.Fatalf("Ошибка при копировании версии в архив: %s\n", err)
					}
				}
			}
			log.Println("Установленная копия перемещена в архив")

			// Теперь мы можем распаковать новую версию на старое место.
			log.Println("Распаковываю обновлённую версию")
			err = files.Unzip(filepath.Join(conf.Local.Path, archname), conf.Local.Path)
			if err != nil {
				// Если не вышло - падаем и ругаемся.
				log.Fatalf("Ошибка при распаковке архива: %s\n", err)
			}

			// Копируем конфиги в папку app1/conf внутри дистрибутива.
			log.Println("Копирую файлы в необходимые директории.")
			err = files.CopyDir(conf.Local.Conf, filepath.Join(conf.Local.Path, "app1", "conf"))
			if err != nil {
				// Если не вышло - паникуем и падаем.
				log.Fatalf("Ошибка при копировании конфигов: %s\n", err)
			}

			// Копируем бинарники в папку app1/server/bin внутри дистрибутива.
			err = files.CopyDir(conf.Local.Bin, filepath.Join(conf.Local.Path, "app1", "server", "bin"))
			if err != nil {
				// Если не вышло - паникуем и падаем.
				log.Fatalf("Ошибка при копировании стартовых файлов: %s\n", err)
			}

			// Проверяем размер и дату файла jruby.jar
			log.Println("Проверка версии jruby.jar")
			copyJruby := false // Нужно ли обновлять jruby.
			// Получаем информацию о файле, используя переменную окружения.
			// TODO: Добавить проверку наличия переменных окружения вообще.
			localJrubyInfo, err := os.Stat(filepath.Join(os.Getenv("JRUBY_HOME"), "lib", "jruby.jar"))
			if err != nil {
				// Ели локального файла нет - значит надо его скопировать, но это странно.
				log.Print("Файл не найден в требуемой директории, принудительно копирую.")
				copyJruby = true
			}
			// Получаем информацию о файле в дистрибутиве.
			newJrubyInfo, err := os.Stat(filepath.Join(conf.Local.Path, "app1", "gems", "jruby.jar"))
			if err != nil {
				// А вот если его там нет - ругаемся и падаем.
				log.Fatal("Jruby.jar не найден в дистрибутиве.")
			}

			// Сравниваем время модификации и размеры файла.
			// TODO: Заменить на сравнение SHA хешей, это корректнее.
			if (localJrubyInfo.ModTime() != newJrubyInfo.ModTime()) || (localJrubyInfo.Size() != newJrubyInfo.Size()) {
				// Если они отличаются - заменяем.
				log.Print("Версии jruby различаются, обновляем локальный.")
				copyJruby = true
			}
			// Ну и при принудительном обновлении - тоже обновляем.
			if conf.System.Force {
				copyJruby = true
			}

			// Собственно, само копирование.
			if copyJruby {
				// Копируем файл из дистрибутива в jruby_home
				err = files.CopyFile(filepath.Join(conf.Local.Path, "app1", "gems", "jruby.jar"), filepath.Join(os.Getenv("jruby_home"), "lib", "jruby.jar"))
				if err != nil {
					// Если скопировать не получилось - ругаемся и падаем.
					log.Fatalf("Ошибка при копировании jruby.jar: %s\n", err)
				}
			}

			// Проверяем наличие необходимых гемов.
			log.Println("Проверка установленных gem")
			// Получаем список локальных гемов
			gemsLocal, err := files.GetGemList()
			if err != nil {
				// Если не вышло - ругаемся и падаем.
				log.Fatalf("Ошибка при проверке установленных gem: %s\n", err)
			}
			// Получаем список гемов из дистрибутива.
			gemsDist, err := files.GetDistGems(conf.Local.Path)
			if err != nil {
				// Если не вышло - ругаемся и палаем.
				log.Fatalf("Ошибка при проверке дистрибутивных gem: %s\n", err)
			}

			forceGems := false // нужно ли обновлять гемы.

			for _, elem := range gemsDist {
				if !contains(gemsLocal, elem) {
					log.Printf("Найден отсутствующий gem: %s. Он будет установлен.\n", elem)
					forceGems = true
				}

			}

			if forceGems {
				installCmd := exec.Command(filepath.Join(conf.Local.Path, "app1", "gems", "install_2.bat"))
				installCmd.Dir = filepath.Join(conf.Local.Path, "app1", "gems")
				installCmd.Stderr = os.Stdout
				err := installCmd.Run()
				if err != nil {
					log.Printf("Установка завершилась с ошибкой: %s\n", err)
				}

			} else {
				log.Printf("Версии установленных gem не требуют обновления.")
			}

		}

	} else {
		log.Println("Обновление не требуется.")
	}

	if conf.FirstRun {
		log.Fatal("Вы запускаете обновление первый раз, скопируйте требуемые файлы в папки с конфигами и стартовыми файлами и запустите обновление повторно.")
	} else {
		if *custUpdateOnly {
			log.Printf("Выбран режим без запуска Вектора. Завершаю работу.")
		} else {
			log.Println("Запуск АРМ.")
			appCmd := exec.Command(filepath.Join(conf.Local.Path, "app1", "server", "bin", conf.Local.Start))
			appCmd.Dir = filepath.Join(conf.Local.Path, "app1", "server", "bin")
			err = appCmd.Start()
			if err != nil {
				log.Print(err)
			}
		}
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
