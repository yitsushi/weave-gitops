package reader

import (
	"io/fs"
	"testing"
	"testing/fstest"
	"time"

	. "github.com/onsi/gomega"
)

type walkerFixture struct {
	*GomegaWithT
}

func arrangeSimpleFS() fstest.MapFS {
	fileOne := &fstest.MapFile{
		Data:    []byte("test 1"),
		Mode:    0,
		ModTime: time.Date(2020, 11, 5, 8, 32, 11, 0, time.UTC),
		Sys:     nil,
	}

	return fstest.MapFS{
		"apps/app-1/test1.txt":                  fileOne,
		"apps/app-1/overlays/staging/test1.txt": fileOne,
		"apps/app-2/test1.txt":                  fileOne,
		"apps/app-2/test2.txt":                  fileOne,
		"apps/app-2/overlays/dev/test1.txt":     fileOne,
		"apps/app-2/overlays/dev/test2.txt":     fileOne,
	}
}

func setUpWalkerTest(t *testing.T) walkerFixture {
	return walkerFixture{
		GomegaWithT: NewGomegaWithT(t),
	}
}

func TestWalkAndRead_Simple(t *testing.T) {
	f := setUpWalkerTest(t)

	var paths []string

	walkDir := WalkDir(&paths)
	err := fs.WalkDir(arrangeSimpleFS(), "apps", walkDir)

	f.Expect(err).To(BeNil())
	f.Expect(paths).To(HaveLen(6))

	var paths2 []string

	walkDir = WalkDir(&paths2)
	err = fs.WalkDir(arrangeSimpleFS(), "apps/app-1", walkDir)

	f.Expect(err).To(BeNil())
	f.Expect(paths2).To(HaveLen(2))

	var paths3 []string

	walkDir = WalkDir(&paths3)
	err = fs.WalkDir(arrangeSimpleFS(), "apps/app-2", walkDir)

	f.Expect(err).To(BeNil())
	f.Expect(paths3).To(HaveLen(4))
}

func TestWalkAndRead_NoPath(t *testing.T) {
	f := setUpWalkerTest(t)

	var paths []string

	walkDir := WalkDir(&paths)
	err := fs.WalkDir(arrangeSimpleFS(), "clusters", walkDir)

	f.Expect(err).To(BeNil())
	f.Expect(paths).To(HaveLen(0))
}

func TestWalkAndRead_OddFormat(t *testing.T) {
	f := setUpWalkerTest(t)

	var paths []string

	walkDir := WalkDir(&paths)
	err := fs.WalkDir(arrangeSimpleFS(), "./app/app-1/", walkDir)

	f.Expect(err).To(BeNil())
	f.Expect(paths).To(HaveLen(0))
}
