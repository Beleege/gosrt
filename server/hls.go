package server

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/beleege/gosrt/config"
	"github.com/beleege/gosrt/core/selector"
	"github.com/beleege/gosrt/protocol/hls"
	"github.com/beleege/gosrt/protocol/mpegts"
	"github.com/beleege/gosrt/protocol/srt"
	"github.com/beleege/gosrt/util/log"
)

const (
	_maxTSDuration = 5.0
)

var (
	_crossdomainxml = `<?xml version="1.0" ?>
<cross-domain-policy>
	<allow-access-from domain="*" />
	<allow-http-request-headers-from domain="*" headers="*"/>
</cross-domain-policy>`
	_tsCache = hls.NewTSCache()
)

func SetupHLSServer() {
	log.Infof("######################## HLS Server start at port:%d #####################", config.GetHLSPort())
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GetHLSPort()))
	if err != nil {
		log.Errorf("hls server fail: %s", err.Error())
		panic(err)
	}
	go start(listener)
	// just get on session for test
	list := selector.GetAllSession()
	for len(list) == 0 {
		// wait session build
		time.Sleep(100 * time.Millisecond)
		list = selector.GetAllSession()
	}

	if len(list) > 0 {
		for i := range list {
			go onData(list[i].RecWin.ListenBatch())
		}
	}
}

func start(listener net.Listener) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handle)
	_ = http.Serve(listener, mux)
}

func handle(w http.ResponseWriter, r *http.Request) {
	log.Debugf("request url is %s", r.URL.Path)
	if path.Base(r.URL.Path) == "crossdomain.xml" {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(_crossdomainxml))
		return
	}

	p := strings.TrimLeft(r.URL.Path, "/")

	switch path.Ext(r.URL.Path) {
	case ".m3u8":
		//key := strings.Split(p, path.Ext(p))[0]
		playlist, err := _tsCache.GetPlayList()
		if err != nil {
			log.Debugf("get playlist error: %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Debugf("************ playlist: %s", string(playlist))
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "application/x-mpegURL")
		w.Header().Set("Content-Length", strconv.Itoa(len(playlist)))
		_, _ = w.Write(playlist)
	case ".ts":
		paths := strings.SplitN(p, "/", 3)
		key := paths[0] + "/" + paths[1]
		item, err := _tsCache.GetItem(key)
		if err != nil {
			log.Debugf("get ts item error: %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "video/mp2ts")
		w.Header().Set("Content-Length", strconv.Itoa(len(item.Data)))
	}
}

func onData(ch chan []*srt.DataPacket) {
	var tsBuf *bytes.Buffer
	first := -1.0
	for data := range ch {
		for i := range data {
			if len(data[i].Content) > 0 {
				buf := bytes.NewBuffer(data[i].Content)
				for b := buf.Next(mpegts.TSPackageSize); len(b) == mpegts.TSPackageSize; b = buf.Next(mpegts.TSPackageSize) {
					if d, ok := mpegts.ExtractPCR(b); ok {
						log.Infof("get ts pcr is %f", d)
						if first < 0 {
							first = d
						} else if d-first > _maxTSDuration && tsBuf != nil {
							// build ts item FIXME replace 'test'
							key := fmt.Sprintf("/%s/%d.ts", "test", time.Now().Unix())
							_tsCache.SetItem(key, data[i].SequenceNum, d-first, tsBuf.Bytes())
							first = d
							tsBuf = bytes.NewBuffer(nil)
						}
					}
					if tsBuf == nil {
						tsBuf = bytes.NewBuffer(nil)
					}
					tsBuf.Write(data[i].Content)
				}
			}
			// reduce gc press
			selector.Recycle(data[i].Content)
			data[i].Content = nil
		}
	}
}
