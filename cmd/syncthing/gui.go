package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/codegangsta/martini"
)

type guiError struct {
	Time  time.Time
	Error string
}

var (
	configInSync = true
	guiErrors    = []guiError{}
	guiErrorsMut sync.Mutex
)

func startGUI(addr string, m *Model) {
	router := martini.NewRouter()
	router.Get("/", getRoot)
	router.Get("/rest/version", restGetVersion)
	router.Get("/rest/model", restGetModel)
	router.Get("/rest/connections", restGetConnections)
	router.Get("/rest/config", restGetConfig)
	router.Get("/rest/config/sync", restGetConfigInSync)
	router.Get("/rest/need", restGetNeed)
	router.Get("/rest/system", restGetSystem)
	router.Get("/rest/errors", restGetErrors)

	router.Post("/rest/config", restPostConfig)
	router.Post("/rest/restart", restPostRestart)
	router.Post("/rest/error", restPostError)

	go func() {
		mr := martini.New()
		mr.Use(embeddedStatic())
		mr.Use(martini.Recovery())
		mr.Action(router.Handle)
		mr.Map(m)
		err := http.ListenAndServe(addr, mr)
		if err != nil {
			warnln("GUI not possible:", err)
		}
	}()
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/index.html", 302)
}

func restGetVersion() string {
	return Version
}

func restGetModel(m *Model, w http.ResponseWriter) {
	var res = make(map[string]interface{})

	globalFiles, globalDeleted, globalBytes := m.GlobalSize()
	res["globalFiles"], res["globalDeleted"], res["globalBytes"] = globalFiles, globalDeleted, globalBytes

	localFiles, localDeleted, localBytes := m.LocalSize()
	res["localFiles"], res["localDeleted"], res["localBytes"] = localFiles, localDeleted, localBytes

	inSyncFiles, inSyncBytes := m.InSyncSize()
	res["inSyncFiles"], res["inSyncBytes"] = inSyncFiles, inSyncBytes

	files, total := m.NeedFiles()
	res["needFiles"], res["needBytes"] = len(files), total

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetConnections(m *Model, w http.ResponseWriter) {
	var res = m.ConnectionStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetConfig(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(cfg)
}

func restPostConfig(req *http.Request) {
	err := json.NewDecoder(req.Body).Decode(&cfg)
	if err != nil {
		log.Println(err)
	} else {
		saveConfig()
		configInSync = false
	}
}

func restGetConfigInSync(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(map[string]bool{"configInSync": configInSync})
}

func restPostRestart(req *http.Request) {
	restart()
}

type guiFile File

func (f guiFile) MarshalJSON() ([]byte, error) {
	type t struct {
		Name string
		Size int64
	}
	return json.Marshal(t{
		Name: f.Name,
		Size: File(f).Size,
	})
}

func restGetNeed(m *Model, w http.ResponseWriter) {
	files, _ := m.NeedFiles()
	gfs := make([]guiFile, len(files))
	for i, f := range files {
		gfs[i] = guiFile(f)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gfs)
}

var cpuUsagePercent float64
var cpuUsageLock sync.RWMutex

func restGetSystem(w http.ResponseWriter) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	res := make(map[string]interface{})
	res["myID"] = myID
	res["goroutines"] = runtime.NumGoroutine()
	res["alloc"] = m.Alloc
	res["sys"] = m.Sys
	cpuUsageLock.RLock()
	res["cpuPercent"] = cpuUsagePercent
	cpuUsageLock.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetErrors(w http.ResponseWriter) {
	guiErrorsMut.Lock()
	json.NewEncoder(w).Encode(guiErrors)
	guiErrorsMut.Unlock()
}

func restPostError(req *http.Request) {
	bs, _ := ioutil.ReadAll(req.Body)
	req.Body.Close()
	showGuiError(string(bs))
}

func showGuiError(err string) {
	guiErrorsMut.Lock()
	guiErrors = append(guiErrors, guiError{time.Now(), err})
	if len(guiErrors) > 5 {
		guiErrors = guiErrors[len(guiErrors)-5:]
	}
	guiErrorsMut.Unlock()
}
