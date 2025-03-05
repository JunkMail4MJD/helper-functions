// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    HARFile, err := UnmarshalHARFile(bytes)
//    bytes, err = HARFile.Marshal()

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"maps"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const DEBUG = false

func UnmarshalHARFile(data []byte) (HARFile, error) {
	var r HARFile
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *HARFile) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type HARFile struct {
	Log *HARLog `json:"log,omitempty"`
}

type HARLog struct {
	Creator *Creator `json:"creator,omitempty"`
	Entries []Entry  `json:"entries,omitempty"`
	Pages   []Page   `json:"pages,omitempty"`
	Version *string  `json:"version,omitempty"`
}

type Creator struct {
	Name    *string `json:"name,omitempty"`
	Version *string `json:"version,omitempty"`
}

type Entry struct {
	FromCache       *string            `json:"_fromCache,omitempty"`
	Pageref         *string            `json:"pageref,omitempty"`
	Request         *Request           `json:"request,omitempty"`
	Response        *Response          `json:"response,omitempty"`
	ServerIPAddress *string            `json:"serverIPAddress,omitempty"`
	StartedDateTime *time.Time         `json:"startedDateTime,omitempty"`
	Time            *float64           `json:"time,omitempty"`
	Timings         map[string]float64 `json:"timings,omitempty"`
}

type Request struct {
	BodySize    *int64    `json:"bodySize,omitempty"`
	Cookies     []Cooky   `json:"cookies,omitempty"`
	Headers     []Header  `json:"headers,omitempty"`
	HeadersSize *int64    `json:"headersSize,omitempty"`
	HTTPVersion *string   `json:"httpVersion,omitempty"`
	Method      *string   `json:"method,omitempty"`
	PostData    *PostData `json:"postData,omitempty"`
	QueryString []Header  `json:"queryString,omitempty"`
	URL         *string   `json:"url,omitempty"`
}

type Cooky struct {
	Domain   *string    `json:"domain,omitempty"`
	Expires  *time.Time `json:"expires"`
	HTTPOnly *bool      `json:"httpOnly,omitempty"`
	Name     *string    `json:"name,omitempty"`
	Path     *string    `json:"path,omitempty"`
	SameSite *string    `json:"sameSite,omitempty"`
	Secure   *bool      `json:"secure,omitempty"`
	Value    *string    `json:"value,omitempty"`
}

type Header struct {
	Name  *string `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

type PostData struct {
	MIMEType *string  `json:"mimeType,omitempty"`
	Params   []Header `json:"params,omitempty"`
	Text     *string  `json:"text,omitempty"`
}

type Response struct {
	Error                   *string  `json:"_error,omitempty"`
	FetchedViaServiceWorker *bool    `json:"_fetchedViaServiceWorker,omitempty"`
	TransferSize            *int64   `json:"_transferSize,omitempty"`
	BodySize                *int64   `json:"bodySize,omitempty"`
	Content                 *Content `json:"content,omitempty"`
	Cookies                 []Cooky  `json:"cookies,omitempty"`
	Headers                 []Header `json:"headers,omitempty"`
	HeadersSize             *int64   `json:"headersSize,omitempty"`
	HTTPVersion             *string  `json:"httpVersion,omitempty"`
	RedirectURL             *string  `json:"redirectURL,omitempty"`
	Status                  *int64   `json:"status,omitempty"`
	StatusText              *string  `json:"statusText,omitempty"`
}

type Content struct {
	Compression *int64                 `json:"compression,omitempty"`
	Encoding    *string                `json:"encoding,omitempty"`
	MIMEType    *string                `json:"mimeType,omitempty"`
	Size        *int64                 `json:"size,omitempty"`
	Text        *string                `json:"text,omitempty"`
	JSON        map[string]interface{} `json:"json,omitempty"`
}

type Page struct {
	ID              *string      `json:"id,omitempty"`
	PageTimings     *PageTimings `json:"pageTimings,omitempty"`
	StartedDateTime *time.Time   `json:"startedDateTime,omitempty"`
	Title           *string      `json:"title,omitempty"`
}

type PageTimings struct {
	OnContentLoad *float64 `json:"onContentLoad,omitempty"`
	OnLoad        *float64 `json:"onLoad,omitempty"`
}

func main() {

	var filenameArg string

	log.Printf("\n**************************************\n\n")
	if len(os.Args) == 2 {
		// Access the first argument after the program name
		filenameArg = os.Args[len(os.Args)-1]
		log.Println("Filename parameter    : ", filenameArg)
	} else {
		log.Fatalln("Usage: splitHAR <filename>  ")
	}

	f := checkFile(filenameArg)
	v, err := os.ReadFile(f) //read the content of file
	check_err(err)

	filehash := bytesToHash(v)

	har_data, err := UnmarshalHARFile(v)
	check_err(err)

	log.Printf("\n**************************************\n")
	log.Println("File hash: ", filehash)
	log.Println("Creator  : ", toJSONStringPretty(har_data.Log.Creator))
	log.Println("Version  : ", toJSONStringPretty(har_data.Log.Version))
	log.Println("Pages    : \n", toJSONStringPretty(har_data.Log.Pages))

	processHAR(har_data, filehash)

	printMemoryUsage()
}

func processHAR(h HARFile, fhash string) {
	if h.Log != nil {
		processLog(h.Log, fhash)
	} else {
		log.Fatalln("HAR file does not have a 'log' object present in file.")
	}
}

func processLog(l *HARLog, fhash string) {
	if l != nil {
		earliestTime := processPages(l.Pages)
		processEntries(l.Entries, fhash, earliestTime)
	} else {
		log.Fatalln("No HAR file does not have a 'log.entries' object to process.")
	}
}

func processPages(p []Page) time.Time {

	earliestTime := time.Now()
	count := len(p)
	log.Println("HAR file has ", count, " pages listed.")

	for i, rec := range p {
		if rec.StartedDateTime != nil {
			time2 := *rec.StartedDateTime
			time2Str := getZuluTime(time2)
			if DEBUG {
				log.Println("page: ", i+1, " of ", count, " start time: ", time2Str)
			}
			earliestTime = keepEarliest(earliestTime, time2)
		}
	}
	return earliestTime
}

func keepEarliest(t1, t2 time.Time) time.Time {

	earliestTime := t1
	if t2.Before(earliestTime) {
		earliestTime = t2
	}
	return earliestTime
}

func getZuluTime(t time.Time) string {
	timebs, err := t.MarshalText()
	check_err(err)
	return string(timebs)
}

func processEntries(e []Entry, fhash string, earliestTime time.Time) {

	count := len(e)

	log.Println("HAR file has ", count, " entries.")

	earliestTime = getEarliestStart(e, earliestTime)

	log.Println("Earliest Start time overall : ", getZuluTime(earliestTime))

	filedate := getZuluTime(earliestTime)
	filedate = filedate[0:10]
	log.Println("file date: ", filedate)

	rFQDNs := make(map[string]int)
	paths := make(map[string]int)
	responseTypes := make(map[string]int)

	for i, rec := range e {
		if DEBUG {
			log.Printf("**************************************")
		}
		if i%10 == 0 {
			log.Println("Processing Entry : ", i+1, " of ", count)
		}

		recordJSON := toJSONStringCompact(rec)
		recordHash := bytesToHash([]byte(recordJSON))

		reverseHost := ""
		path := ""
		method := ""
		mimeType := ""

		req, err := extractRequest(rec)
		if err == nil {
			reverseHost, path = extractRequestURL(req)
			path = strings.ReplaceAll(path, "~", "_my_")
			path = strings.ReplaceAll(path, "{refId}", "_refId_")
			method = extractRequestMethod(req)
		}

		res, err := extractResponse(rec)
		if err == nil {
			mimeType = extractResponseMimeType(res)
		}

		rFQDNs[reverseHost] += 1
		paths[path] += 1
		responseTypes[mimeType] += 1
		if DEBUG {
			log.Println("Extracted request method: ", method)
			log.Println("Extracted reverse hostname: ", reverseHost, " URI Path: ", path)
			log.Println("Extracted response MIME Type: ", mimeType)
			if i >= 5 {
				log.Println("Exiting due to index meeting or exceeding limit of 5: ", i)
				break
			}
		}

		recordType := "assets"
		if mimeType == "application/json" {
			recordType = "data"
		}
		dir := "date=" + filedate
		dir += "/rev_domain=" + reverseHost
		dir += "/type=" + recordType
		if recordType == "data" && !strings.Contains(path, ",") {

			if strings.Contains(path, ".") {
				webdir := filepath.Dir(path)
				dir = filepath.Join(dir, webdir)
			} else {
				dir = filepath.Join(dir, path)
			}

			body, err := parseResponseContent(res)
			if err == nil {
				recordJSON, _ = sjson.Set(recordJSON, "response.content.json", body)
				var strPointer *string
				recordJSON, _ = sjson.Set(recordJSON, "response.content.text", strPointer)
			}
		}

		if DEBUG {
			log.Println("Create Directory: ", dir)
		}

		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatalln(err)
		}
		recordJSON, err = sjson.Set(recordJSON, "file_hash", fhash)
		check_err(err)

		recordJSON, err = sjson.Set(recordJSON, "record_hash", recordHash)
		check_err(err)

		aFile := filepath.Join(dir, "entries.json")
		appendEntryToFile(aFile, recordJSON+"\n")
	}

	log.Println("Aggregate request domains: ")
	printMap(rFQDNs)

	log.Println("Aggregate request paths: ")
	printMap(paths)

	log.Println("Aggregate response types: ")
	printMap(responseTypes)
}

func parseResponseContent(res Response) (map[string]interface{}, error) {

	if res.Content != nil {
		c := *res.Content
		if DEBUG {
			log.Printf("response content: type: %T value: %v \n", c, c)
		}

		if c.Text != nil {
			text := *c.Text
			if DEBUG {
				log.Printf("response content.text : type: %T value: %v \n", text, text)
			}
			body := make(map[string]interface{})
			err := json.Unmarshal([]byte(text), &body)
			if err == nil {
				if DEBUG {
					log.Printf("response content.text parsed by json.unmarshal. type: %T value: %v \n", body, body)
				}

				if DEBUG {
					jString, _ := sjson.Set("", "body", body)
					log.Printf("response content body jString. type: %T value: %v \n", jString, jString)

					value := gjson.Get(jString, "body|@pretty:{\"sortKeys\":true}")

					log.Printf("response gjson content. type: %T value: %v \n", value, value)
					//os.Exit(1)
				}
				return body, nil

			} else {
				return make(map[string]interface{}), err
			}
		}
	}
	return make(map[string]interface{}), errors.New("Response Content is null.")
}

func getEarliestStart(e []Entry, earliest time.Time) time.Time {

	earliestStartTime := earliest
	if DEBUG {
		log.Printf("**************************************")
	}
	count := len(e)
	for i, rec := range e {

		t, err := extractStartTime(rec)

		var s string
		if err != nil {
			s = ""
		} else {
			earliestStartTime = keepEarliest(earliestStartTime, t)
			s = getZuluTime(t)
		}
		if DEBUG {
			log.Println("Processing Start time for Entry : ", i+1, " of ", count, " start time: ", s)
		}
	}
	return earliestStartTime
}

func appendEntryToFile(fn, json string) {
	f, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln("failed to open file: ", fn, " Error: ", err)
	}
	defer f.Close()
	if _, err := f.WriteString(json); err != nil {
		log.Fatalln("failed to append to file: ", fn, " Error: ", err)
	}
}

func printMap(m map[string]int) {

	log.Printf("\t count \t\t value\n")
	log.Printf("\t ===== \t\t =====\n")
	keys := make([]string, 0, len(m))
	for k := range maps.Keys(m) {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for _, k := range keys {
		log.Printf("\t %v \t\t %v\n", m[k], k)
	}
}

func extractResponseMimeType(r Response) string {
	if r.Content != nil {
		con := r.Content
		if con.MIMEType != nil {
			result := *con.MIMEType
			return strings.ToLower(result)
		}
	}
	return ""
}

func extractResponse(r Entry) (Response, error) {
	if r.Response != nil {
		response := *r.Response
		return response, nil
	}
	var r2 Response
	return r2, errors.New("HAR file log.entry[_] object is missing a response object.")
}

func extractRequest(r Entry) (Request, error) {
	if r.Request != nil {
		request := *r.Request
		return request, nil
	}
	var r2 Request
	return r2, errors.New("HAR file log.entry[_] object is missing a request object.")
}

func extractRequestMethod(r Request) string {
	if r.Method != nil {
		result := *r.Method
		return strings.ToLower(result)
	}
	return ""
}

func extractStartTime(r Entry) (time.Time, error) {

	if r.StartedDateTime != nil {
		t := *r.StartedDateTime
		return t, nil
	}
	return time.Now(), errors.New("Start time is missing")
}

func extractRequestURL(req Request) (string, string) {
	var reverseHost, path string

	if req.URL != nil {
		url := *req.URL
		reverseHost, path = parseUrl(url)
	} else {
		log.Fatalln("HAR file log.entry[_].request object is missing a URL key.")
	}

	return strings.ToLower(reverseHost), path
}

func check_err(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}

func toJSONStringCompact(t interface{}) string {
	bs, err := json.Marshal(t)
	check_err(err)
	return string(bs)
}

func toJSONStringPretty(t interface{}) string {
	//bs, err := json.Marshal(t)
	bs, err := json.MarshalIndent(t, "", "  ")
	check_err(err)
	return string(bs)
}

func bytesToHash(bs []byte) string {
	h := sha256.New()
	h.Write(bs)
	b := h.Sum(nil)
	hexString := hex.EncodeToString(b)

	return hexString
}

func isValidPath(p string) bool {
	_, err := filepath.Abs(p)
	return err == nil
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return !os.IsNotExist(err)
}

func hasReadPermission(p string) bool {
	fileInfo, err := os.Stat(p)
	if err != nil {
		return false
	}
	return fileInfo.Mode().Perm()&0400 != 0
}

func sanitizePath(p string) string {
	t, err := filepath.Abs(p)
	check_err(err)
	if DEBUG {
		log.Println("Absolute path         : ", t)
	}
	return filepath.Clean(p)
}

func checkFile(p string) string {

	sanitizedPath := sanitizePath(p)

	if !isValidPath(sanitizedPath) {
		log.Fatalln("Invalid path syntax")
	}

	if !fileExists(sanitizedPath) {
		log.Fatalln("File does not exist")
	}

	if !hasReadPermission(sanitizedPath) {
		log.Fatalln("Insufficient permissions")
	}

	return sanitizedPath
}

func printMemoryUsage() {

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	alloc := fmt.Sprintf("%12.3f", float64(memStats.Alloc)/1024.0/1024.0)
	tot := fmt.Sprintf("%12.3f", float64(memStats.TotalAlloc)/1024.0/1024.0)
	sys := fmt.Sprintf("%12.3f", float64(memStats.Sys)/1024.0/1024.0)
	heap := fmt.Sprintf("%12.3f", float64(memStats.HeapAlloc)/1024.0/1024.0)
	heapsys := fmt.Sprintf("%12.3f", float64(memStats.HeapSys)/1024.0/1024.0)
	log.Printf("**************************************")
	log.Println("Alloc      (MB): ", alloc)   // Total bytes allocated
	log.Println("TotalAlloc (MB): ", tot)     // Total bytes allocated (even if freed)
	log.Println("Sys        (MB): ", sys)     // Total bytes of memory obtained from the OS
	log.Println("HeapAlloc  (MB): ", heap)    // Bytes allocated on the heap
	log.Println("HeapSys    (MB): ", heapsys) // Bytes of heap memory obtained from the OS
	log.Printf("**************************************")

}

func parseUrl(s string) (string, string) {

	u, err := url.Parse(s)
	if err != nil {
		log.Fatal("Error Parsing URL:", err)
		panic(err)
	}

	if DEBUG {
		log.Println("Original  :", s)
		log.Println("Scheme    :", u.Scheme)

		log.Println("User      :", u.User)
		log.Println("  Username:", u.User.Username())
		p, _ := u.User.Password()
		log.Println("  Password:", p)
	}
	hostString := u.Host
	if DEBUG {
		log.Println("Host      :", u.Host)
	}

	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		if DEBUG {
			log.Println("Error parsing host string: ", err)
		}
	} else {
		hostString = host
		if DEBUG {
			log.Println("  host    :", host)
			log.Println("  port    :", port)
		}
	}

	reverseHost := getReverseFQDN(hostString)
	if DEBUG {
		log.Println("Host str  :", hostString)
		log.Println("reverse Host : ", reverseHost)
		log.Println("Path         : ", u.Path)
		log.Println("Fragment  :", u.Fragment)
		log.Println("RawQuery  :", u.RawQuery)
		m, _ := url.ParseQuery(u.RawQuery)
		log.Println("Parsed Query: ", m)
		log.Println("theKey?     : ", m["theKey"][0])
	}

	return reverseHost, u.Path
}

func getReverseFQDN(fqdn string) string {
	result := strings.Split(fqdn, ".")
	slices.Reverse(result)
	return strings.Join(result, ".")
}
