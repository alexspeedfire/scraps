package files

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

//ZipFiles рекурсивно упаковывает указанную в basePath папку в архив
func ZipFiles(w *zip.Writer, basePath, baseInZip string) error {
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			dat, err := ioutil.ReadFile(basePath + file.Name())
			if err != nil {
				return err
			}

			f, err := w.Create(baseInZip + file.Name())
			if err != nil {
				return err
			}
			_, err = f.Write(dat)
			if err != nil {
				return err
			}
		} else if file.IsDir() {
			newBase := basePath + file.Name() + "/"
			ZipFiles(w, newBase, baseInZip+file.Name()+"/")
		}
	}
	return nil
}

//Unzip распаковывает файл src по адресу dest
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	extractAndWrite := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWrite(f)
		if err != nil {
			return err
		}
	}
	return nil
}

//CopyFile копирует файл из src в dst
func CopyFile(src, dst string) error {
	srcfd, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcfd.Close()

	dstfd, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstfd.Close()

	_, err = io.Copy(dstfd, srcfd)
	if err != nil {
		return err
	}

	srcinfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

//CopyDir копирует содержимое src в dst.
func CopyDir(src, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	srcinfo, err = os.Stat(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, srcinfo.Mode())
	if err != nil {
		return err
	}

	fds, err = ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			err = CopyDir(srcfp, dstfp)
			if err != nil {
				return err
			}

		} else {
			err = CopyFile(srcfp, dstfp)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//GetGemList возвращает список гемов, установленных в системе.
func GetGemList() ([]string, error) {

	cmd := exec.Command("gem", "list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	for i := range lines {
		lines[i] = strings.Split(lines[i], " ")[0]
	}
	sort.Strings(lines)
	return lines, nil
}

//GetDistGems возвращает список гемов в дистрибутиве
func GetDistGems(rootpath string) ([]string, error) {

	installScript := filepath.Join(rootpath, "app1", "gems", "install_2.bat")
	if _, err := os.Stat(installScript); os.IsNotExist(err) {
		return nil, err
	}
	buff, err := ioutil.ReadFile(installScript)
	if err != nil {
		return nil, err
	}
	installSrc := strings.Fields(string(buff))

	installGems := make([]string, 0)
	for i := range installSrc {
		if strings.HasPrefix(installSrc[i], "'.\\vendor") {
			//installSrc = append(installSrc[:i], installSrc[i+1:]...)
			newstring := strings.TrimPrefix(installSrc[i], "'.\\vendor\\cache\\")
			newstring = strings.TrimSuffix(newstring, ".gem'")
			newstring = strings.Split(newstring, ".")[0]
			tmp := len(newstring)
			newstring = newstring[:tmp-2]
			//fmt.Println(newstring)
			installGems = append(installGems, newstring)
			//installSrc = installSrc[:i+copy(installSrc[i:], installSrc[i+1:])]

		}
		//		log.Printf("%d %s\n", i, installSrc[i])
	}
	sort.Strings(installGems)
	return installGems, nil
}
