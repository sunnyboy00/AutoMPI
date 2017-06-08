package AutoMPI

import (
	"fmt"
	"net/http"
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

	fmt.Fprintf(w, "<!DOCTYPE html><html><head><title>AutiMPI Node: %s</title><meta http-equiv=\"refresh\" content=\"5\"><style>body {font-family: monospace;}</style></head><body></br>\n", base.MyNodeGUID)

	fmt.Fprintf(w, "*****************************   AutoMPI   *****************************</br>\n")
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintf(w, "NodeGUID: %s</br>\n", base.MyNodeGUID)
	fmt.Fprintf(w, "Communications port: %s</br>\n", base.LocalAddressString+receivingPort)
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintf(w, "*******************************************************************</br>\n")
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintln(w, "Number of local workers:", len(base.workers), "</br>")
	fmt.Fprintf(w, "*************************** Worker List ***************************</br>\n")
	for key, value := range base.workers {
		fmt.Fprintf(w, "%s  Age: %s s</br>\n", key, value.GetAge())
	}
	fmt.Fprintf(w, "*************************** *********** ***************************</br>\n")
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintln(w, "************************ Outgoing links:", len(base.OutgoingLinks), "************************</br>")
	for key, value := range base.OutgoingLinks {
		fmt.Fprintf(w, "NodeGUID: %s Address: %s Age:%s seconds &nbsp;<a href=\"http://%s%s\">Visit</a> </br> \n", key, value.GetRemoteAddressAndPortAsString(), value.GetAge(), value.GetRemoteAddressAsString(), hTTPStatePort)
	}
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintf(w, "</br>\n")
	fmt.Fprintln(w, "************************ Incoming links:", len(base.IncommingLinks), "************************</br>")
	for key, value := range base.IncommingLinks {
		fmt.Fprintf(w, "NodeGUID: %s Address: %s Age:%s seconds &nbsp;<a href=\"http://%s%s\">Visit</a> </br> \n", key, value.GetRemoteAddressAndPortAsString(), value.GetAge(), value.GetRemoteAddressAsString(), hTTPStatePort)
	}

	fmt.Fprintln(w, "</body></html>")
}
