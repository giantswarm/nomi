package unit

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

type BeaconObserver struct {
	unitEngine *UnitEngine
}

func NewBeaconObserver(engine *UnitEngine) *BeaconObserver {
	return &BeaconObserver{
		unitEngine: engine,
	}
}

func (s *BeaconObserver) StartHTTPService(addr string) {
	r := mux.NewRouter()
	r.HandleFunc("/hello/{beaconID}", withIDParam(s.HelloHandler)).Methods("GET")
	r.HandleFunc("/alive/{beaconID}", withIDParam(s.AliveHandler)).Methods("GET")
	r.HandleFunc("/bye/{beaconID}", withIDParam(s.ByeHandler)).Methods("GET")

	r.HandleFunc("/stats/{statsID}", s.StatsHandler).Methods("POST")

	http.Handle("/", r)
	glog.V(3).Infof("listening on %s\n", addr)
	go http.ListenAndServe(addr, nil)
	glog.V(2).Infof("listening on %s\n", addr)
}

func withIDParam(handler func(beaconID string, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		beaconID := mux.Vars(r)["beaconID"]
		if beaconID == "" {
			glog.Errorln("empty beaconID")
			w.WriteHeader(400)
		} else {
			handler(beaconID, w, r)
		}
	}
}

func (s *BeaconObserver) HelloHandler(beaconID string, w http.ResponseWriter, r *http.Request) {
	delay := s.unitEngine.MarkUnitRunning(beaconID)
	glog.V(2).Infof("marked beacon as running: %s [%d] %f", beaconID, len(s.unitEngine.runningUnits), delay.Seconds())
	w.Write([]byte("ok.\n"))
}

func (s *BeaconObserver) AliveHandler(beaconID string, w http.ResponseWriter, r *http.Request) {
	if _, isStopped := s.unitEngine.stoppingUnits[beaconID]; isStopped {
		w.WriteHeader(500)
	} else {
		//	glog.V(2).Infoln("ensuring beacon is marked as running: " + beaconID)
		w.Write([]byte("ok.\n"))
	}
}
func (s *BeaconObserver) ByeHandler(beaconID string, w http.ResponseWriter, r *http.Request) {
	glog.V(2).Infof("marking beacon as stopped: %s [%d]", beaconID, len(s.unitEngine.stoppedUnits)+len(s.unitEngine.runningUnits)+len(s.unitEngine.startingUnits))
	s.unitEngine.MarkUnitStopped(beaconID)
	w.Write([]byte("ok.\n"))
}

func (s *BeaconObserver) StatsHandler(w http.ResponseWriter, r *http.Request) {
	statsID := mux.Vars(r)["statsID"]
	b := bytes.NewBufferString("")
	b.ReadFrom(r.Body)

	var hostname string
	var cpuusage float64
	var rss int

	n, err := fmt.Sscanf(b.String(), "%s %f %d", &hostname, &cpuusage, &rss)
	if err != nil || n != 3 {
		glog.Warningln("don't know how to parse statsline: " + b.String())
		return
	}

	s.unitEngine.DumpProcessStats(statsID, hostname, cpuusage, rss)
}
