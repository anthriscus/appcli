package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/anthriscus/appcli/filer"
)

// data repository
// potential to expand to load from json database/mongodb/sql datastore

var (
	sessionDatabase TodoListItems
)

func IsOpen() bool {
	return (sessionDatabase != nil)
}

func Open(ctx context.Context, storageFile string) error {
	list, err := Restore(ctx, storageFile)
	if err != nil {
		fmt.Printf("Fat; error restoring list file err: %s storageFile: %s\n", err, storageFile)
		return err
	}
	sessionDatabase = list
	// if sessionDatabase != nil {

	// }
	return nil
}

// restore from json file
func Restore(ctx context.Context, storageFile string) (TodoListItems, error) {
	// destination, err := os.OpenFile(storageFile, openFlag, readwriteFileMode)
	destination, err := filer.OpenFileRestore(storageFile)
	if err != nil {
		fmt.Printf("Error restoring list file err: %s storageFile: %s\n", err, storageFile)
		// errorLogger.Log.ErrorContext(ctx, "Error restoring list file", "err", err, "storageFile", storageFile)
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
		// errorLogger.Log.ErrorContext(ctx, "Error restoring data", "err", err)
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
			fmt.Println("returning empty list json error")
			fmt.Printf("Error restoring list from json err:%s\n", err)
			// errorLogger.Log.ErrorContext(ctx, "Error restoring list from json", "err", err)
			return TodoListItems{}, nil
		}
		return restoredList, nil
	}
}

// save list back to json file
func Save(ctx context.Context, storageFile string, list TodoListItems) error {

	if data, err := json.Marshal(list); err != nil {
		fmt.Printf("Save failed converting todo list to json err:%s\n", err)
		// errorLogger.Log.ErrorContext(ctx, "Save failed converting todo list to json", "err", err)
		return err
	} else {
		// if destination, err := os.OpenFile(storageFile, openTruncateFlag, readwriteFileMode); err != nil {
		if destination, err := filer.OpenFileTruncate(storageFile); err != nil {
			fmt.Printf("Save failed getting file err:%s storageFile: %s\n", err, storageFile)
			// errorLogger.Log.ErrorContext(ctx, "Save failed getting file", "err", err, "storageFile", storageFile)
			return err
		} else {
			defer destination.Close()
			if _, err := destination.Write(data); err != nil {
				fmt.Printf("Save to file failed err:%s storageFile:%s\n", err, storageFile)
				// errorLogger.Log.ErrorContext(ctx, "Save to file failed ", "err", err, "storageFile", storageFile)
				return err
			}
		}
	}
	fmt.Printf("Saved data storageFile:%s\n", storageFile)
	// ActivityLogger.Log.InfoContext(ctx, "Saved data", "storageFile", storageFile)
	return nil
}
