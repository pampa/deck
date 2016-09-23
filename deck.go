package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/boltdb/bolt"
	"os"
	"path/filepath"
	"sort"
)

type Deck struct {
	Root   string
	Data   string
	Prune  []string
	Ignore []string
	db     *bolt.DB
}

var picks = []byte("picks")
var index = []byte("index")

var deck Deck

func (d *Deck) Init(f string) {
	if _, err := toml.DecodeFile(f, &d); err != nil {
		log.Error(f, ":", err)
	}
	log.Debug("Root", d.Root)
	log.Debug("Data", d.Data)
	log.Debug("Prune", d.Prune)
	log.Debug("Ignore", d.Ignore)

	if _, err := os.Stat(d.Root); err != nil {
		log.Error(err)
	}

	if err := os.MkdirAll(d.Data, 0744); err != nil {
		log.Error(err)
	}

	var err error
	if d.db, err = bolt.Open(d.Data+"/deck.db", 0644, nil); err != nil {
		log.Error(err)
	}

	if err := d.db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(index); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(picks); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error(err)
	}
}

func (d *Deck) Close() {
	d.db.Close()
}

func (d *Deck) Scan(hash bool, pick bool) {

	var newFiles []string
	var pickedFiles []string
	var modifiedFiles []string

	if err := d.db.View(func(tx *bolt.Tx) error {
		bkPicks := tx.Bucket(picks)
		bkIndex := tx.Bucket(index)
		filepath.Walk(d.Root, func(p string, i os.FileInfo, _ error) error {
			if i.IsDir() && matchAny(p, d.Prune) {
				log.Debug("Prune", p)
				return filepath.SkipDir
			} else if matchAny(p, d.Ignore) {
				log.Debug("Ignore", p)
				return nil
			} else {
				if i.Mode().IsRegular() || (i.Mode()&os.ModeSymlink != 0) {
					pk := bkPicks.Get([]byte(p))
					kn := bkIndex.Get([]byte(p))
					if pk != nil {
						pickedFiles = append(pickedFiles, p)
					} else if kn != nil {
						//log.Debug("Know", p)
						fs := getFileObject(p, hash)
						fk := readFileObject(kn)
						if err := fs.IsDifferent(fk, hash); err != nil {
							log.Debug(err, p)
							modifiedFiles = append(modifiedFiles, p)
						}
					} else {
						newFiles = append(newFiles, p)
					}
				} else {
					//log.Debug("Skip", p)
				}
				return nil
			}
		})
		return nil
	}); err != nil {
		log.Error(err)
	}

	if pick {
		deck.Pick(newFiles)
		deck.Pick(modifiedFiles)
		pickedFiles = append(pickedFiles, newFiles...)
		pickedFiles = append(pickedFiles, modifiedFiles...)
		newFiles = nil
		modifiedFiles = nil
	}

	printFiles("New files", newFiles)
	printFiles("Modified files", modifiedFiles)
	printFiles("Picked files", pickedFiles)
}

func (d *Deck) Pick(files []string) {
	d.db.Update(func(tx *bolt.Tx) error {
		bkPicks := tx.Bucket(picks)
		for _, f := range files {
			s, err := os.Lstat(f)
			if os.IsNotExist(err) {
				log.Error(err)
			}
			if s.Mode().IsRegular() || (s.Mode()&os.ModeSymlink != 0) {
				log.Debug("pick", f)
				if err := bkPicks.Put([]byte(f), nil); err != nil {
					return err
				}
			} else {
				log.Error("Only regular files and symlinks allowed", f)
			}
		}
		return nil
	})
}

func (d *Deck) Unpick(all bool, files []string) {
	d.db.Update(func(tx *bolt.Tx) error {
		bkPicks := tx.Bucket(picks)

		if all {
			bkPicks.ForEach(func(k, v []byte) error {
				log.Debug("unpick", string(k))
				if err := bkPicks.Delete(k); err != nil {
					return err
				}
				return nil
			})
		} else {
			for _, f := range files {
				log.Debug("unpick", f)
				if err := bkPicks.Delete([]byte(f)); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (d *Deck) Remove(files []string) {
	d.db.Update(func(tx *bolt.Tx) error {
		bkIndex := tx.Bucket(index)
		for _, f := range files {
			log.Debug("remove", f)
			if err := bkIndex.Delete([]byte(f)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *Deck) Reset(files []string) {
	d.db.View(func(tx *bolt.Tx) error {
		bkIndex := tx.Bucket(index)
		for _, f := range files {
			kn := bkIndex.Get([]byte(f))
			if kn == nil {
				log.Error("File not in index", f)
			}
			fo := readFileObject(kn)
			fo.Reset(f)
		}
		return nil
	})
}

func (d *Deck) Commit(pak string, ver string) {
	d.db.Update(func(tx *bolt.Tx) error {
		bkPicks := tx.Bucket(picks)
		bkIndex := tx.Bucket(index)

		bkPicks.ForEach(func(k, v []byte) error {
			log.Debug("commit", pak, ver, string(k))
			fo := getFileObject(string(k), true)
			fo.Package = Package{
				Name:    pak,
				Version: ver,
			}

			if err := bkIndex.Put(k, fo.ToBytes()); err != nil {
				log.Error(err)
			}

			fo.Stov(string(k))

			if err := bkPicks.Delete(k); err != nil {
				log.Error(err)
			}

			return nil
		})
		return nil
	})
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

func appendPackage(s []Package, n Package) []Package {
	for _, i := range s {
		if i == n {
			return s
		}
	}
	return append(s, n)
}

func (d *Deck) Packages() []Package {
	var packages []Package
	d.db.View(func(tx *bolt.Tx) error {
		bkIndex := tx.Bucket(index)
		bkIndex.ForEach(func(k, v []byte) error {
			fo := readFileObject(v)
			packages = appendPackage(packages, fo.Package)
			return nil
		})
		return nil
	})
	sort.Sort(ByName(packages))
	return packages
}

func (d *Deck) Show(pak string, mk bool) {
	d.db.View(func(tx *bolt.Tx) error {
		bkIndex := tx.Bucket(index)
		bkIndex.ForEach(func(k, v []byte) error {
			fo := readFileObject(v)
			if fo.Package.Name == pak {
				fmt.Println(string(k))
			}
			return nil
		})
		return nil
	})
}

func (d *Deck) List(ver bool) {
	for _, p := range d.Packages() {
		if ver {
			fmt.Println(p.Name)
		} else {
			fmt.Println(p.Name, p.Version)
		}
	}
}
