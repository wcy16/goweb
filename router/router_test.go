// test and benchmark for 1000 urls access

package router

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

var alphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// generate url of length (1 - n) of given times
func generator() func(n int, times int) string {

	history := make(map[string]bool)

	return func(n int, times int) string {
		ret := ""
		for {
			l := rand.Intn(n) + 1
			for t := 0; t != times; t++ {
				b := make([]rune, l)
				for i := range b {
					b[i] = alphabet[rand.Intn(len(alphabet))]
				}
				ret += "/" + string(b)
			}
			if history[ret] {
				ret = ""
			} else {
				history[ret] = true
				break
			}
		}
		return ret
	}
}

var generateURL = generator()

var route = NewRoute()
var table = make([]testTable, 0, 800)

type testTable struct {
	method  string
	url     string
	content string
}

// use this function to prevent bugs caused by closure
func generateHandler(response string) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprintf(writer, response)
	}
}

func init() {
	var url string
	// create route table
	prefix := "static"
	for i := 0; i != 50; i++ {
		// depth 2 - 5
		for j := 1; j != 5; j++ {
			url = prefix + generateURL(10, j)
			route.GET(url, generateHandler(url))
			table = append(table, testTable{"GET", url, url})
			url = prefix + generateURL(10, j)
			route.POST(url, generateHandler(url))
			table = append(table, testTable{"POST", url, url})
		}
	}

	static := len(table)

	// dynamic
	prefix = "dynamic"
	var t1, t2 string
	for i := 0; i != 10; i++ {
		// one parameter
		t1 = generateURL(10, 1)
		t2 = generateURL(10, 1)
		route.POST(prefix+t1+"/#id"+t2, func(writer http.ResponseWriter, request *http.Request) {
			param, ok := request.Context().Value("param").(map[string]string)
			if !ok {
				_, _ = fmt.Fprintf(writer, "wrong")
			} else {
				_, _ = fmt.Fprintf(writer, "/"+param["id"])
			}
		})
		for j := 0; j != 20; j++ {
			randID := generateURL(10, 1)
			table = append(table, testTable{"POST", fmt.Sprintf("%s%s%s%s", prefix, t1, randID, t2), randID})
		}

		// two parameters
		t1 = generateURL(10, 1)
		t2 = generateURL(10, 1)
		route.GET(prefix+t1+"/#two1"+t2+"/#two2", func(writer http.ResponseWriter, request *http.Request) {
			param, ok := request.Context().Value("param").(map[string]string)
			if !ok {
				_, _ = fmt.Fprintf(writer, "wrong")
			} else {
				_, _ = fmt.Fprintf(writer, "/"+param["two1"]+"/"+param["two2"])
			}
		})
		for j := 0; j != 20; j++ {
			rand1 := generateURL(10, 1)
			rand2 := generateURL(10, 1)
			table = append(table, testTable{"GET", fmt.Sprintf("%s%s%s%s%s", prefix, t1, rand1, t2, rand2), rand1 + rand2})
		}

		// three parameters
		t1 = generateURL(10, 1)
		route.GET(prefix+t1+"/#t1"+"/#t2"+"/#t3", func(writer http.ResponseWriter, request *http.Request) {
			param, ok := request.Context().Value("param").(map[string]string)
			if !ok {
				_, _ = fmt.Fprintf(writer, "wrong")
			} else {
				_, _ = fmt.Fprintf(writer, "/"+param["t1"]+"/"+param["t2"]+"/"+param["t3"])
			}
		})
		for j := 0; j != 20; j++ {
			rand1 := generateURL(10, 1)
			rand2 := generateURL(10, 1)
			rand3 := generateURL(10, 1)
			table = append(table, testTable{"GET", fmt.Sprintf("%s%s%s%s%s", prefix, t1, rand1, rand2, rand3), rand1 + rand2 + rand3})
		}
	}
	total := len(table)
	dynamic := total - static
	fmt.Printf("total %d urls, %d static and %d dynamic\n", total, static, dynamic)
}

func TestRoute(t *testing.T) {
	for _, element := range table {
		recorder := httptest.NewRecorder()
		req, err := http.NewRequest(element.method, element.url, nil)
		if err != nil {
			t.Fatal(err)
		}
		route.ServeHTTP(recorder, req)
		if recorder.Body.String() != element.content {
			t.Errorf("unexpected return: got %v want %v at %v, method: %v",
				recorder.Body.String(), element.content, element.url, element.method)
		}
	}
}

type mockWriter struct{}

func (m *mockWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockWriter) WriteHeader(int) {}

func BenchmarkRoute(b *testing.B) {
	//b.ReportAllocs()
	//b.ResetTimer()
	writer := new(mockWriter)
	for i := 0; i < b.N; i++ {
		for _, element := range table {
			req, _ := http.NewRequest(element.method, element.url, nil)
			route.ServeHTTP(writer, req)
		}
	}
}
