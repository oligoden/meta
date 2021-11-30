package entity_test

import (
	"fmt"
	"testing"

	"github.com/oligoden/meta/entity"
	"github.com/stretchr/testify/assert"
)

func TestEncNil(t *testing.T) {
	assert := assert.New(t)

	b := &entity.Branch{}
	result, err := b.Build(nil)
	assert.EqualError(err, "encountered nil")
	assert.Nil(result)
	assert.Empty(b.Filename)
	assert.Empty(b.Directories)
	assert.Equal([]string{}, b.Directories)
}

func TestFileEncNil(t *testing.T) {
	assert := assert.New(t)
	e := &entity.File{
		Name: "a",
	}

	b := &entity.Branch{}
	result, err := b.Build(e)
	assert.EqualError(err, "encountered nil")
	assert.Nil(result)
	assert.Equal("a", b.Filename)
	assert.Empty(b.Directories)
	assert.Equal([]string{}, b.Directories)
}

func TestDirEncNil(t *testing.T) {
	assert := assert.New(t)
	e := &entity.Directory{
		Basic: entity.Basic{
			Name: "a",
		},
	}

	b := &entity.Branch{}
	result, err := b.Build(e)
	assert.EqualError(err, "encountered nil")
	assert.Nil(result)
	assert.Empty(b.Filename)
	assert.Equal([]string{"a"}, b.Directories)
}

func TestFileDirEncNil(t *testing.T) {
	assert := assert.New(t)
	e := &entity.File{
		Name: "a",
		Parent: &entity.Directory{
			Basic: entity.Basic{
				Name: "a",
			},
		},
	}

	b := &entity.Branch{}
	result, err := b.Build(e)
	assert.EqualError(err, "encountered nil")
	assert.Nil(result)
	assert.Equal([]string{"a"}, b.Directories)
}

func TestFileInProject(t *testing.T) {
	assert := assert.New(t)
	e := &entity.File{
		Name:   "a",
		Parent: &entity.Project{},
	}

	b := &entity.Branch{}
	result, err := b.Build(e)
	assert.NoError(err)
	assert.Equal(&entity.Project{}, result)
}

func TestFileDirProject(t *testing.T) {
	assert := assert.New(t)
	e := &entity.File{
		Name: "a",
		Parent: &entity.Directory{
			Basic: entity.Basic{
				Name:   "a",
				Parent: &entity.Project{},
			},
		},
	}

	b := &entity.Branch{}
	result, err := b.Build(e)
	assert.NoError(err)
	assert.Equal(&entity.Project{}, result)
	assert.Equal("a", b.Filename)
	assert.Equal([]string{"a"}, b.Directories)
}

func TestFileDirsProject(t *testing.T) {
	assert := assert.New(t)
	e := &entity.File{
		Name: "a",
		Parent: &entity.Directory{
			Basic: entity.Basic{
				Name: "a",
				Parent: &entity.Directory{
					Basic: entity.Basic{
						Name:   "b",
						Parent: &entity.Project{},
					},
				},
			},
		},
	}

	b := &entity.Branch{}
	result, err := b.Build(e)
	assert.NoError(err)
	assert.Equal(&entity.Project{}, result)
	assert.Equal("a", b.Filename)
	assert.Equal([]string{"a", "b"}, b.Directories)
}

func TestDepthExceeded(t *testing.T) {
	assert := assert.New(t)
	e := &entity.Directory{
		Basic: entity.Basic{
			Name: "dir",
		},
	}

	for i := 0; i < 39; i++ {
		e = &entity.Directory{
			Basic: entity.Basic{
				Name:   fmt.Sprint("dir", i),
				Parent: e,
			},
		}
	}

	b := &entity.Branch{}
	_, err := b.Build(e)
	assert.EqualError(err, "branch depth exceeded")
	fmt.Println(b.Directories)
	assert.Len(b.Directories, 40)
}

func TestProjectBranchEncNil(t *testing.T) {
	assert := assert.New(t)

	b := &entity.ProjectBranch{}
	result, err := b.Build(nil)
	assert.EqualError(err, "building default branch -> encountered nil")
	assert.Nil(result)
	assert.Empty(b.Filename)
	assert.Empty(b.Directories)
	assert.Equal([]string{}, b.Directories)
}

func TestProjectBranchWithFileDirProject(t *testing.T) {
	assert := assert.New(t)
	e := &entity.File{
		Name: "a",
		Parent: &entity.Directory{
			Basic: entity.Basic{
				Name:   "a",
				Parent: &entity.Project{},
			},
		},
	}

	b := &entity.ProjectBranch{}
	result, err := b.Build(e)
	assert.NoError(err)
	assert.Equal(&entity.Project{}, result)
	assert.Equal("a", b.Filename)
	assert.Equal([]string{"a"}, b.Directories)
}
