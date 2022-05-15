package plugin

import (
	"context"
	"reflect"

	//"encoding/base64"
	"encoding/json"
	"net/http"

	//"strconv"

	//"math/rand"
	"time"

	"sort"

	"github.com/buger/jsonparser"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
	"golang.org/x/net/websocket"
)

// Make sure SampleDatasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler, backend.StreamHandler interfaces. Plugin should not
// implement all these interfaces - only those which are required for a particular task.
// For example if plugin does not need streaming functionality then you are free to remove
// methods that implement backend.StreamHandler. Implementing instancemgmt.InstanceDisposer
// is useful to clean up resources used by previous datasource instance when a new datasource
// instance created upon datasource settings changed.
var (
	_ backend.QueryDataHandler      = (*SampleDatasource)(nil)
	_ backend.CheckHealthHandler    = (*SampleDatasource)(nil)
	_ backend.StreamHandler         = (*SampleDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*SampleDatasource)(nil)
)

// NewSampleDatasource creates a new datasource instance.
func NewSampleDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &SampleDatasource{}, nil
}

// SampleDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type SampleDatasource struct{}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *SampleDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		log.DefaultLogger.Info("This is the Query ", q)
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type qModel struct {
	queryText     string `json:"queryText"`
	WithStreaming bool   `json:"withStreaming"`

	//constant      int    `json:"constant"`
}

var st = ""

//var stop = make(chan bool)

//var que string = ""

func (d *SampleDatasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}
	//log.DefaultLogger.Info("query is ", query.JSON)
	//stop <- false
	qu, err := jsonparser.GetString(query.JSON, "queryText")
	//que.fin := string(que.quer)
	//que := string(quer)
	st = string(qu)

	if err != nil {
		log.DefaultLogger.Info("JsonParser error is", err)
	}
	log.DefaultLogger.Info("This is the st string", st)
	log.DefaultLogger.Info("This is the value")

	//som := jsonMap["data"]
	//log.DefaultLogger.Info("This is json Map", st)
	// keys := make([]string, 0, len(jsonMap["data"]))
	// for k := range jsonMap["data"] {
	// 	keys = append(keys, k)
	// }
	// sort.Strings(keys)
	// log.DefaultLogger.Info(keys[0])

	// for _, k := range keys {
	// 	//fmt.Println(k, population[k])
	// 	log.DefaultLogger.Info("Value in the array is ", k, jsonMap["data"][k])

	// }
	// dat, err := base64.StdEncoding.DecodeString(string(val))
	// if err != nil {

	// }

	//log.DefaultLogger.Info("This is the val string after decoding", dat)
	// Unmarshal the JSON into our queryModel.
	var qm qModel

	response.Error = json.Unmarshal(query.JSON, &qm)
	log.DefaultLogger.Info("Error is", response.Error)
	if response.Error != nil {
		return response
	}

	log.DefaultLogger.Info("Qm is ", qm)

	// create data frame response.
	frame := data.NewFrame("response")

	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, make([]time.Time, 1)),
		data.NewField("values", nil, make([]string, 1)),
	)
	frame.Fields[1].Set(0, st)

	// If query called with streaming on then return a channel
	// to subscribe on a client-side and consume updates from a plugin.
	// Feel free to remove this if you don't need streaming for your datasource.
	if qm.WithStreaming {
		channel := live.Channel{
			Scope:     live.ScopeDatasource,
			Namespace: pCtx.DataSourceInstanceSettings.UID,
			Path:      query.QueryType,
		}
		frame.SetMeta(&data.FrameMeta{Channel: channel.String()})
	}
	/*if query.QueryType == "new" {
		log.DefaultLogger.Info("Inside the if Function")
		stop <- true
	}*/
	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *SampleDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

// SubscribeStream is called when a client wants to connect to a stream. This callback
// allows sending the first message.
func (d *SampleDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream called", "request", req)

	status := backend.SubscribeStreamStatusOK
	//if req.Path == "stream" {
	// Allow subscribing only on expected path.
	//	status = backend.SubscribeStreamStatusOK
	//}
	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel
func DoneAsync(val chan []byte) {
	//r := make(chan []byte, 5096)
	log.DefaultLogger.Info("This is the quer string", st)
	log.DefaultLogger.Info("Length of st is", len(st))
	//if len(st) == 0 {
	//	st = "ws://localhost:8080/vui/platforms/volttron1/pubsub/devices/Campus/Building1/Fake1/all"
	//}
	config, err := websocket.NewConfig(st, st)
	if err != nil {
		log.DefaultLogger.Error(err.Error())
	}

	config.Header = http.Header{
		"Authorization": {"Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJncm91cHMiOlsiYWRtaW4iXSwiaWF0IjoxNjM3MDIyMzE2LCJuYmYiOjE2MzcwMjIzMTYsImV4cCI6MTYzNzAyMzIxNiwiZ3JhbnRfdHlwZSI6ImFjY2Vzc190b2tlbiJ9.V7WhUBwEc0rBq4jAtKbeWizSJErr5nVMd6kGcoh4k9g"},
	}
	ws, err := websocket.DialConfig(config)
	if err != nil {
		log.DefaultLogger.Error(err.Error())
	}
	var n int
	retMsg := make([]byte, 5096)
	for {
		n, err = ws.Read(retMsg)
		//log.DefaultLogger.Info("This is the retMsg", retMsg)
		val <- retMsg[:n]
		//log.DefaultLogger.Info("This is the response", string(<-val))

	}
	//if err != nil {
	//log.DefaultLogger.Info(string(retMsg[:n]))
	//}

	//log.DefaultLogger.Info(string(retMsg[:n]))
	//r <- retMsg[:n]
	//log.DefaultLogger.Info("This is the method", string(<-r))
	//return r

}

func (d *SampleDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream called", "request", req)

	// Create the same data frame as for query data.
	//frame := data.NewFrame("response")

	// Add fields (matching the same schema used in QueryData).
	//frame.Fields = append(frame.Fields,
	//data.NewField("time", nil, make([]time.Time, 1)),
	//data.NewField("value", nil, make([]string, 1)),
	/*data.NewField("OutsideAirTemperature1", nil, make([]string, 1)),
	data.NewField("SampleLong1", nil, make([]string, 1)),
	data.NewField("SampleBool1", nil, make([]string, 1)),
	data.NewField("OutsideAirTemperature2", nil, make([]string, 1)),
	data.NewField("SampleLong2", nil, make([]string, 1)),
	data.NewField("SampleBool2", nil, make([]string, 1)),
	data.NewField("OutsideAirTemperature3", nil, make([]string, 1)),
	data.NewField("SampleLong3", nil, make([]string, 1)),
	data.NewField("SampleBool3", nil, make([]string, 1)),
	data.NewField("Heartbeat", nil, make([]string, 1)),
	data.NewField("EKG", nil, make([]string, 1)),
	data.NewField("SampleWritableFloat1", nil, make([]string, 1)),
	data.NewField("SampleWritableShort1", nil, make([]string, 1)),
	data.NewField("SampleWritableFloat2", nil, make([]string, 1)),
	data.NewField("SampleWritableShort2", nil, make([]string, 1)),
	data.NewField("SampleWritableBool2", nil, make([]string, 1)),
	data.NewField("SampleWritableFloat3", nil, make([]string, 1)),
	data.NewField("SampleWritableShort3", nil, make([]string, 1)),
	data.NewField("SampleWritableBool3", nil, make([]string, 1)),
	data.NewField("PowerState", nil, make([]string, 1)),
	data.NewField("ValveState", nil, make([]string, 1)),
	data.NewField("EKG_Sin", nil, make([]string, 1)),
	data.NewField("EKG_Cos", nil, make([]string, 1)),*/
	//)
	//counter := 0
	val := make(chan []byte, 5096)
	go DoneAsync(val)
	//log.DefaultLogger.Info("This is the val variable", string(<-val))
	// Stream data frames periodically till stream closed by Grafana.
	for {
		select {
		/*case abc := <-stop:
		if abc == true {
			ct, cancel := context.WithCancel(ctx)
			log.DefaultLogger.Info(ct.Err().Error())
			cancel()
		}*/
		case <-ctx.Done():
			log.DefaultLogger.Info("Context done, finish streaming", "path", req.Path)
			return nil
		case al := <-val:
			frame := data.NewFrame("response")
			frame.Fields = append(frame.Fields,
				data.NewField("time", nil, make([]time.Time, 1)))
			// Send new data periodically.
			//log.DefaultLogger.Info("This is the strn variable", string(al))
			stringAl := string(al)
			var jsonMap []map[string]interface{}
			errr := json.Unmarshal([]byte(stringAl), &jsonMap)
			log.DefaultLogger.Info("this is the error", errr)
			log.DefaultLogger.Info("This is json", jsonMap[0])
			log.DefaultLogger.Info("This is string al", stringAl)
			var i = 1
			frame.Fields[0].Set(0, time.Now())
			keys := make([]string, 0, len(jsonMap[0]))
			for k := range jsonMap[0] {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			log.DefaultLogger.Info(keys[0])

			for _, k := range keys {
				//fmt.Println(k, population[k])
				log.DefaultLogger.Info("Value in the array is ", k, jsonMap[0][k])

			}
			//for key, element := range jsonMap[0] {
			for _, k := range keys {
				log.DefaultLogger.Info(k)
				//log.DefaultLogger.Info(string(element))
				//str, ok := element.(string)
				//if ok {
				//	log.DefaultLogger.Info("This is a string", str)
				//}
				//log.DefaultLogger.Info("type of value", reflect.TypeOf(element).String())

				var element = jsonMap[0][k]

				if reflect.TypeOf(element).String() == "bool" {
					frame.Fields = append(frame.Fields,
						data.NewField(string(k), nil, make([]bool, 1)))
					frame.Fields[i].Set(0, true)
					i++
					log.DefaultLogger.Info("This is inside bool i", string(i))
				}
				if reflect.TypeOf(element).String() == "float64" {
					frame.Fields = append(frame.Fields,
						data.NewField(string(k), nil, make([]float64, 1)))

					frame.Fields[i].Set(0, element)
					i++
					log.DefaultLogger.Info("This is inside float i", string(i))
				}
				//}

			}
			i = 1
			/*for j := range jsonMap[0] {
				log.DefaultLogger.Info(j)
				log.DefaultLogger.Info("Indi", jsonMap[0][j])
				frame.Fields = append(frame.Fields, data.NewField(j, nil, jsonMap[0][j]))
			}*/
			//			frame.Fields = append(frame.Fields, data.NewField("test", nil, make([]string, 1)))
			//			frame.Fields[1].Set(0, string(al))
			//			frame.Fields[2].Set(0, "test")

			//log.DefaultLogger.Info("this is al", string(al))
			//counter++
			log.DefaultLogger.Info("This is the frame", frame)
			err := sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}
		}
	}
}

// PublishStream is called when a client sends a message to the stream.
func (d *SampleDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}
