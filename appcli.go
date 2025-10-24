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
	logFileName           string = "todolistserver.log"
)

type runmode int

const (
	RunModeCLI int = iota
	RunModeServer
)

// var ServerLogger logging.AppLogger
var runMode runmode

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

	runMode = runmode(RunModeCLI)

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

	// wire up logger
	logName := dir + "\\" + logFileName
	if logFileHandle, err := filer.OpenLogFile(logName); err == nil {
		defer logFileHandle.Close()
		logOptions := logging.LoggerOptions()
		logging.Setup(logFileHandle, logOptions)
		logging.Log().InfoContext(ctx, "Starting up logging with static logger")
	}

	// init / pickup current list before process command
	storageFile := fmt.Sprintf("%s\\%s", dir, dataFileName)
	// open the database for cli and api
	openErr := store.OpenSession(ctx, storageFile)
	if openErr != nil {
		// fatal database is unavailable
		return
	}

	// process the flags
	switch {
	case *flagAdd != "":
		if nextKey, ok := store.AddTask(ctx, *flagAdd); ok == nil {
			store.ListTask(nextKey)
		}
	case *flagUpdate > 0 && len(taskDescription) > 0:
		if ok := store.DescriptionChange(ctx, *flagUpdate, taskDescription); ok == nil {
			store.ListTask(*flagUpdate)
		}
	case *flagNotStart > 0:
		if ok := store.StateChange(ctx, *flagNotStart, store.StateNotStarted); ok == nil {
			store.ListTask(*flagNotStart)
		}
	case *flagStart > 0:
		if ok := store.StateChange(ctx, *flagStart, store.StateStarted); ok == nil {
			store.ListTask(*flagStart)
		}
	case *flagComplete > 0:
		if ok := store.StateChange(ctx, *flagComplete, store.StateCompleted); ok == nil {
			store.ListTask(*flagComplete)
		}
	case *flagDelete > 0:
		if ok := store.DeleteTask(ctx, *flagDelete); ok == nil {
			store.ListTask(-1)
		}
	case *flagList:
		store.ListTask(taskId)
	case *flagRunServer:
		runMode = runmode(RunModeServer)
		api.Run()
	}

	if runMode == runmode(RunModeCLI) {
		// write back to the file
		store.SaveSession(ctx, storageFile)
	}
}
