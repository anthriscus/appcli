package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strconv"

	"github.com/anthriscus/appcli/api"
	"github.com/anthriscus/appcli/appcontext"
	"github.com/anthriscus/appcli/filer"
	"github.com/anthriscus/appcli/logging"
	"github.com/anthriscus/appcli/store"
)

const (
	dataStorageFolderName string = "appcli"
	dataFileName          string = "todolist.json"
	errorFileName         string = "appcliError.log"
	activityFileName      string = "appcliActivity.log"
	serverLogFileName     string = "todolistserver.log"
)

var errorLogger logging.AppLogger
var ActivityLogger logging.AppLogger
var ServerLogger logging.AppLogger

func main() {
	// input flags
	var flagAdd = flag.String("add", "", "add todolist item (\"description\")")
	var flagUpdate = flag.Int("update", 0, "update task item description (id -description \"new description\")")
	var flagNotStart = flag.Int("notstart", 0, "set task item id number to not started ( id )")
	var flagStart = flag.Int("start", 0, "start a task item ( id )")
	var flagComplete = flag.Int("complete", 0, "complete a task item id ( id )")
	var flagDelete = flag.Int("delete", 0, "delete a task item id number ( id )")
	var flagList = flag.Bool("list", false, "list items in the todolist item ( with additional optional -taskid num to show one item)")
	var flagRunServer = flag.Bool("runserver", false, "run todolist as http server")

	// additional flag required for description updates
	var taskDescription string
	flag.Func("description", "use this with -update for the update description text -description \"new text\"", func(s string) error {
		if len(s) == 0 {
			return errors.New("value of description needs to be supplied")
		} else {
			taskDescription = s
		}
		return nil
	})
	// additional flag for list filter
	var taskId int
	flag.Func("taskid", "optional, use this -taskid with -list for one task", func(s string) error {
		if i, ok := strconv.Atoi(s); ok != nil {
			return errors.New("value of taskid needs to be supplied")
		} else {
			taskId = i
		}
		return nil
	})

	// // grab the flag input state from command line
	flag.Parse()

	// but TODO "github.com/google/uuid" will provide a better one
	id := store.GenerateId()
	ctx := context.WithValue(context.Background(), appcontext.TraceIdKey, id)

	// resolve the appdata data sub folder
	dir, err := filer.CreateAppDataFolder(dataStorageFolderName)
	if err != nil {
		// don't have a file logger yet!
		fmt.Printf("Error:%s", "Cannot establish working data folder")
		return
	}

	// wire up loggers
	errorLogName := dir + "\\" + errorFileName
	if errorFile, err := filer.OpenLogFile(errorLogName); err == nil {
		defer errorFile.Close()
		errorLogoptions := logging.ErrorOptions()
		errorLogger.Log = logging.SetupLogger(errorFile, errorLogoptions)
	}
	activityLogName := dir + "\\" + activityFileName
	if activityFile, err := filer.OpenLogFile(activityLogName); err == nil {
		defer activityFile.Close()
		activityLogoptions := logging.ActivityOptions()
		ActivityLogger.Log = logging.SetupLogger(activityFile, activityLogoptions)
	}
	serverLogName := dir + "\\" + serverLogFileName
	if serverLog, err := filer.OpenLogFile(serverLogName); err == nil {
		defer serverLog.Close()
		serverLogoptions := logging.ServerOptions()
		ServerLogger.Log = logging.SetupLogger(serverLog, serverLogoptions)
	}

	// init / pickup current list before process command
	storageFile := fmt.Sprintf("%s\\%s", dir, dataFileName)
	todoList, err := store.Restore(ctx, storageFile)
	if err != nil {
		// fatal database is unavailable
		return
	}

	// for the api
	openErr := store.Open(ctx, storageFile)
	if openErr != nil {
		// fatal database is unavailable
		return
	}

	// process the flags
	switch {
	case *flagAdd != "":
		nextKey := store.AddTask(ctx, todoList, *flagAdd)
		store.ListTask(todoList, nextKey)
	case *flagUpdate > 0 && len(taskDescription) > 0:
		store.DescriptionChange(ctx, todoList, *flagUpdate, taskDescription)
		store.ListTask(todoList, *flagUpdate)
	case *flagNotStart > 0:
		store.StateChange(ctx, todoList, *flagNotStart, store.StateNotStarted)
		store.ListTask(todoList, *flagNotStart)
	case *flagStart > 0:
		store.StateChange(ctx, todoList, *flagStart, store.StateStarted)
		store.ListTask(todoList, *flagStart)
	case *flagComplete > 0:
		store.StateChange(ctx, todoList, *flagComplete, store.StateCompleted)
		store.ListTask(todoList, *flagComplete)
	case *flagDelete > 0:
		store.DeleteTask(ctx, todoList, *flagDelete)
		store.ListTask(todoList, -1)
	case *flagList:
		store.ListTask(todoList, taskId)
	case *flagRunServer:
		api.Run(ServerLogger)
	}

	// write back to the file
	store.Save(ctx, storageFile, todoList)
}
