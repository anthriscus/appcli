package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/anthriscus/appcli/filer"
	"github.com/anthriscus/appcli/logging"
)

var (
	sessionDatabase TodoListItems
	datastoreFile   string
)

func IsOpen() bool {
	return (sessionDatabase != nil)
}

func Commit(ctx context.Context) {
	if IsOpen() {
		SaveSession(ctx, datastoreFile)
	}
}

func OpenSession(ctx context.Context, storageFile string) error {
	list, err := Restore(ctx, storageFile)
	if err != nil {
		fmt.Printf("Fatal error restoring list file err: %s storageFile: %s\n", err, storageFile)
		logging.Log().ErrorContext(ctx, "Fatal error restoring list file err", "err", err, "storageFile", storageFile)
		return err
	}
	sessionDatabase = list
	datastoreFile = storageFile
	return nil
}

// save list back to json file
func SaveSession(ctx context.Context, storageFile string) error {

	if data, err := json.Marshal(sessionDatabase); err != nil {
		fmt.Printf("Save failed converting todo list to json err:%s\n", err)
		logging.Log().ErrorContext(ctx, "Save failed converting todo list to json", "err", err)
		return err
	} else {
		if destination, err := filer.OpenFileTruncate(storageFile); err != nil {
			fmt.Printf("Save failed getting file err:%s storageFile: %s\n", err, storageFile)
			logging.Log().ErrorContext(ctx, "Save failed getting file", "err", err, "storageFile", storageFile)
			return err
		} else {
			defer destination.Close()
			if _, err := destination.Write(data); err != nil {
				fmt.Printf("Save to file failed err:%s storageFile:%s\n", err, storageFile)
				logging.Log().ErrorContext(ctx, "Save to file failed ", "err", err, "storageFile", storageFile)
				return err
			}
		}
	}
	fmt.Printf("Saved data storageFile:%s\n", storageFile)
	logging.Log().InfoContext(ctx, "Saved data", "storageFile", storageFile)
	return nil
}

// restore from json file
func Restore(ctx context.Context, storageFile string) (TodoListItems, error) {
	// destination, err := os.OpenFile(storageFile, openFlag, readwriteFileMode)
	destination, err := filer.OpenFileRestore(storageFile)
	if err != nil {
		fmt.Printf("Error restoring list file err: %s storageFile: %s\n", err, storageFile)
		logging.Log().ErrorContext(ctx, "Error restoring list file", "err", err, "storageFile", storageFile)
		return TodoListItems{}, err
	}
	if destination != nil {
		defer destination.Close()
	}
	return restoreList(ctx, destination)
}

func restoreList(ctx context.Context, destination io.Reader) (TodoListItems, error) {
	if restored, err := io.ReadAll(destination); err != nil {
		fmt.Println(err)
		fmt.Printf("error restoring data\n")
		fmt.Printf("Error restoring data err:%s\n", err)
		logging.Log().ErrorContext(ctx, "Error restoring data", "err", err)
		return TodoListItems{}, err
	} else if len(restored) == 0 {
		// not neccessarily an error
		fmt.Printf("returning empty list restored empty\n")
		return TodoListItems{}, nil
	} else {
		data := []byte(string(restored))
		restoredList := TodoListItems{}
		err := json.Unmarshal(data, &restoredList)
		if err != nil {
			fmt.Println(err)
			logging.Log().ErrorContext(ctx, "Error restoring list from json", "err", err)
			return TodoListItems{}, err
		}
		return restoredList, nil
	}
}

// save list back to json file
func Save(ctx context.Context, storageFile string, list TodoListItems) error {

	if data, err := json.Marshal(list); err != nil {
		fmt.Printf("Save failed converting todo list to json err:%s\n", err)
		logging.Log().ErrorContext(ctx, "Save failed converting todo list to json", "err", err)
		return err
	} else {
		if destination, err := filer.OpenFileTruncate(storageFile); err != nil {
			fmt.Printf("Save failed getting file err:%s storageFile: %s\n", err, storageFile)
			logging.Log().ErrorContext(ctx, "Save failed getting file", "err", err, "storageFile", storageFile)
			return err
		} else {
			defer destination.Close()
			if _, err := destination.Write(data); err != nil {
				fmt.Printf("Save to file failed err:%s storageFile:%s\n", err, storageFile)
				logging.Log().ErrorContext(ctx, "Save to file failed ", "err", err, "storageFile", storageFile)
				return err
			}
		}
	}
	fmt.Printf("Saved data storageFile:%s\n", storageFile)
	logging.Log().InfoContext(ctx, "Saved data", "storageFile", storageFile)
	return nil
}
