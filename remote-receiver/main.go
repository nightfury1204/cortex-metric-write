package remote_receiver

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/gorilla/mux"
	"github.com/prometheus/prometheus/prompb"
)

// https://prometheus.io/docs/instrumenting/exposition_formats/

func UnixMillisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func OutputData(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("err : %s", err)
		return
	}
	raw, err := snappy.Decode(nil, data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("err : %s", err)
		return
	}

	var req prompb.WriteRequest

	err = proto.Unmarshal(raw, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("err : %s", err)
		return
	}

	w.WriteHeader(http.StatusOK)

	fmt.Println(req.String())

	fmt.Println()
	fmt.Println()
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/receive", OutputData)
	if err := http.ListenAndServe(":8088", r); err != nil {
		log.Fatal(err)
	}
}
