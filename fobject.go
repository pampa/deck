package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

type Package struct {
	Name    string
	Version string
}

type ByName []Package

func (a ByName) Len() int {
	return len(a)
}

func (a ByName) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}

func (a ByName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type FileObject struct {
	FileMode os.FileMode
	Size     int64
	Package  Package
	Ref      string
	Sha1     []byte
}

func (f FileObject) Reset(dst string) {
	if f.IsLink() {
		_, err := os.Lstat(dst)
		if os.IsNotExist(err) {
		} else {
			if err := os.Remove(dst); err != nil {
				log.Error(err)
			}
		}
		if err := os.Symlink(f.Ref, dst); err != nil {
			log.Error(err)
		}
	} else {
		f.cp(f.objFile(), dst)
		if err := os.Chmod(dst, f.FileMode); err != nil {
			log.Error(err)
		}
	}
}

func (f FileObject) Stov(src string) {
	if !f.IsLink() {
		if err := os.MkdirAll(f.objDir(), 0744); err != nil {
			log.Error(err)
		}
		f.cp(src, f.objFile())
	}
}

func (f FileObject) IsLink() bool {
	return f.FileMode&os.ModeSymlink != 0
}

func (f FileObject) objFile() string {
	return f.objDir() + "/" + hex.EncodeToString(f.Sha1)
}

func (f FileObject) objDir() string {
	return deck.Data + "/" + hex.EncodeToString(f.Sha1)[:2]
}

func (f FileObject) cp(src string, dst string) {
	log.Debug("cp", src, dst)
	srcFile, err := os.Open(src)
	if err != nil {
		log.Error(err)
	}
	dstFile, err := os.Create(dst)
	if err != nil {
		log.Error(err)
	}
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		log.Error(err)
	}
	dstFile.Sync()
	srcFile.Close()
	dstFile.Close()
}

func (f FileObject) IsDifferent(fn FileObject, hash bool) error {
	if f.FileMode != fn.FileMode {
		return errors.New("Mode does not match")
	}
	if f.IsLink() {
		if f.Ref != fn.Ref {
			return errors.New("Ref does not match")
		}
	} else {
		if f.Size != fn.Size {
			return errors.New("Size does not match")
		}

		if hash {
			if bytes.Compare(f.Sha1, fn.Sha1) != 0 {
				return errors.New("Sha1 does not match")
			}
		}
	}
	return nil
}

func (f FileObject) ToBytes() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(f)
	if err != nil {
		log.Error(err)
	}
	return buf.Bytes()
}

func readFileObject(v []byte) FileObject {
	buf := bytes.NewBuffer(v)
	var fo FileObject
	enc := gob.NewDecoder(buf)
	err := enc.Decode(&fo)
	if err != nil {
		log.Error(err)
	}
	return fo
}

func getFileObject(f string, hash bool) FileObject {
	fi, err := os.Lstat(f)
	if err != nil {
		log.Error(err)
	}

	fo := FileObject{
		FileMode: fi.Mode(),
		Size:     fi.Size(),
	}
	if fo.IsLink() {
		fo.Ref, err = os.Readlink(f)
		if err != nil {
			log.Error(err)
		}
	} else {
		if hash {
			h := sha1.New()
			fh, err := os.Open(f)
			if err != nil {
				log.Error(err)
			}
			io.Copy(h, fh)
			fo.Sha1 = h.Sum(nil)
		}
	}
	return fo
}
