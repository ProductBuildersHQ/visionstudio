package store_test

import (
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store/storetest"
)

func TestMemStoreRenameRepository(t *testing.T) {
	storetest.RenameRepository(t, store.NewMemStore())
}

func TestMemStoreRenameRepositoryFinishesAPartialRename(t *testing.T) {
	storetest.RenameRepositoryFinishesAPartialRename(t, store.NewMemStore())
}
