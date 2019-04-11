package config

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

//ConfigName Путь к конфигу по умолчанию.
//const ConfigName string = "config.yml"

//Configuration хранит конфигурацию приложения
type Configuration struct {
	Version string
	Remote  struct {
		Host string
		Path string
		User string
		Pass string
	}
	Local struct {
		Path  string
		Start string
		Conf  string
		Bin   string
	}
	Update struct {
		Log      string
		Archive  string
		Name     string `yaml:"name_format"`
		Compress bool
		Hash     bool
	}
	System struct {
		Force   bool
		Shift   float64
		Timeout int
	}
	FirstRun bool
}

//ReadConfig читает конфиг и возвращает конфигурацию сервера.
func (config *Configuration) ReadConfig(ConfigName string) error {
	configfile, err := ioutil.ReadFile(ConfigName)
	if err != nil {
		return errors.New("Невозможно прочитать файл настроек")
	}
	err = yaml.Unmarshal(configfile, &config)
	if err != nil {
		return errors.New("Ошибка при разборе файла настроек")
	}
	if config.Version != "v1" {
		return errors.New("Неподдерживаемая версия файла настроек")
	}
	if config.Remote.Host == "" {
		return errors.New("Не задан адрес сервера")
	}
	if config.Remote.Path == "" {
		return errors.New("Не задан путь к дистрибутиву")
	}
	if config.Remote.User == "" {
		return errors.New("Не задан логин от сервера")
	}
	if config.Remote.Pass == "" {
		return errors.New("Не задан пароль от сервера")
	}
	if config.Local.Path == "" {
		return errors.New("Не задан путь к APP1 на ПК")
	}
	if config.Local.Conf == "" {
		return errors.New("Не задана папка с настройками APP1")
	}

	if config.Local.Bin == "" {
		return errors.New("Не задана папка со стартовыми файлами APP1")
	}

	if config.Local.Start == "" {
		return errors.New("Не задан стартовый файл")
	}

	if config.Update.Log == "" {
		return errors.New("Не задана папка с логами модуля обновления")
	}

	if config.Update.Archive == "" {
		return errors.New("Не задана папка для резервной копии")
	}

	return nil
}
