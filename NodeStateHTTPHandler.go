package AutoMPI

import (
	"fmt"
	"net/http"
	"runtime"
)

func (base *Node) stateServe() {
	for {

		err := http.ListenAndServe(hTTPStatePort, http.HandlerFunc(base.ServeHTTP))
		if err != nil {

		}
	}
}

func (base *Node) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintf(w, "hello, you've hit %s\n", r.URL.Path)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	//log.Printf("\nAlloc = %v\nTotalAlloc = %v\nSys = %v\nNumGC = %v\n\n", m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC)

	fmt.Fprintf(w, "<!DOCTYPE html><html><head><title>AutiMPI Node: %s</title><meta http-equiv=\"refresh\" content=\"5\"><style>body {font-family: monospace;}</style></head><body></br>\n", base.GUID)

	fmt.Fprintf(w, "*****************************   AutoMPI   *****************************</br>\n")
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintf(w, "NodeGUID: %s</br>\n", base.GUID)
	fmt.Fprintf(w, "Communications port: %s</br>\n", base.localAddressString+receivingPort)
	fmt.Fprintf(w, "Alloc = %vKB TotalAlloc = %vKB Sys = %vKB NumGC = %v </br>\n", m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC)
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintf(w, "*******************************************************************</br>\n")
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintln(w, "Number of local Workers:", len(base.Workers), "</br>")
	fmt.Fprintf(w, "*************************** Worker List ***************************</br>\n")
	for key, value := range base.Workers {
		fmt.Fprintf(w, "%s  Age: %s s</br>\n", key, value.GetAge())
	}
	fmt.Fprintf(w, "*************************** *********** ***************************</br>\n")
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintln(w, "************************ Outgoing links:", len(base.outgoingLinks), "************************</br>")
	for key, value := range base.outgoingLinks {
		fmt.Fprintf(w, "NodeGUID: %s Address: %s Age:%s seconds &nbsp;<a href=\"http://%s%s\">Visit</a> </br> \n", key, value.GetRemoteAddressAndPortAsString(), value.GetAge(), value.GetRemoteAddressAsString(), hTTPStatePort)
	}
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintln(w, "************************ Incoming links:", len(base.incommingLinks), "************************</br>")
	for key, value := range base.incommingLinks {
		fmt.Fprintf(w, "NodeGUID: %s Address: %s Age:%s seconds &nbsp;<a href=\"http://%s%s\">Visit</a> </br> \n", key, value.GetRemoteAddressAndPortAsString(), value.GetAge(), value.GetRemoteAddressAsString(), hTTPStatePort)
	}

	fmt.Fprintln(w, "</body></html>")
}
