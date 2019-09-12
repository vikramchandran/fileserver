package main

import (
	"flag"
	"fmt"
	"github.com/61c-teach/sp19-proj5-userlib"
	"net/http"
	"log"
	"strings"
	_ "strings"
	"time"
)

// This is the handler function which will handle every request other than cache specific requests.

// func rungetFile(ch chan response *, file string) {
// 	ch <- getFile(file)
// }


func handler(w http.ResponseWriter, r *http.Request) {
	// FIXME This should be using the cache!
	// Note that we will be using userlib.ReadFile we provided to read files on the system.
	// The path to the file is given by r.URL.Path and will be the path to the string.
	// Make sure you properly sanitise it (more described in get file).
	/*** MODIFY THIS CODE ***/
	// Reads the file from the disk
	filename := r.URL.Path
	//newfilename := filename
	//ch := make(chan * response)
	//println("About to run getfile!")
	requestresponse := getFile(filename)
	println("Finished getfile!")
	
	// cacherror := cacheresponse.responseError
	
	// response, err := userlib.ReadFile(workingDir, filename)
	// if err != nil {
	// 	// If we have an error from the read, will return the generic file error message and set the error code to follow that.
	// 	http.Error(w, userlib.FILEERRORMSG, userlib.FILEERRORCODE)
	// 	return
	// }

	isfilerror := 0
	istimeouterror := 0

	if requestresponse.responseError != nil {
		if requestresponse.responseError.Error() == "FileError exists" {
			println("Finalized file error")
			http.Error(w, userlib.FILEERRORMSG, userlib.FILEERRORCODE)
			isfilerror = 1
			return
		}

		if requestresponse.responseError.Error() == "TimeOut Error exists" {
			println("Finalized timeout error")
			http.Error(w, userlib.TimeoutString, userlib.TIMEOUTERRORCODE)
			istimeouterror = 1
			return
		}
	}

	// This will automatically set the right content type for the reply as well.

	//Should now be the sanitzed filname
	w.Header().Set(userlib.ContextType, userlib.GetContentType(requestresponse.filename))
	// We need to set the correct header code for a success since we should only succeed at this point.
	if (isfilerror == 0 && istimeouterror == 0) {
		w.WriteHeader(userlib.SUCCESSCODE) // Make sure you write the correct header code so that the tests do not fail!
	// Write the data which is given to us by the response.
	} 
	//w.Write(response)
	println("About to finish handler")
	w.Write(requestresponse.responseData)
}

// This function will handle the requests to acquire the cache status.
// You should not need to edit this function.
func cacheHandler(w http.ResponseWriter, r *http.Request) {
	// Sets the header of the request to a plain text format since we are just dumping information about the cache.
	// Note that we are just putting a fake filename which will get the correct content type.
	w.Header().Set(userlib.ContextType, userlib.GetContentType("cacheStatus.txt"))
	// Set the success code to the proper success code since the action should not fail.
	w.WriteHeader(userlib.SUCCESSCODE)
	// Get the cache status string from the getCacheStatus function.
	w.Write([]byte(getCacheStatus()))
}

// This function will handle the requests to clear/restart the cache.
// You should not need to edit this function.
func cacheClearHandler(w http.ResponseWriter, r *http.Request) {
	// Sets the header of the request to a plain text format since we are just dumping information about the cache.
	// Note that we are just putting a fake filename which will get the correct content type.
	w.Header().Set(userlib.ContextType, userlib.GetContentType("cacheClear.txt"))
	// Set the success code to the proper success code since the action should not fail.
	w.WriteHeader(userlib.SUCCESSCODE)
	// Get the cache status string from the getCacheStatus function.
	w.Write([]byte(CacheClear()))
}

// The structure used for responding to file requests.
// It contains the file contents (if there is any)
// or the error returned when accessing the file.
// Note that it is only used by you so you do not
// need to use all of the fields in it.
type fileResponse struct {
	filename string
	responseData []byte
	responseError error
	responseChan chan *fileResponse
}

// To request files from the cache, we send a message that 
// requests the file and provides a channel for the return
// information.
// Note that it is only used by you so you do not
// need to use all of the fields in it.
type fileRequest struct {
	filename string
	response chan *fileResponse
}

// DO NOT CHANGE THESE NAMES OR YOU WILL NOT PASS THE TESTS
// Port of the server to run on


var port int
// Capacity of the cache in Bytes
var capacity int
// Timeout for file reads in Seconds.
var timeout int
// The is the working directory of the server
var workingDir string

// The channel to pass file read requests to. This is how you will get a file from the cache.
var fileChan = make(chan *fileRequest)
// The channel to pass a request to get back the capacity info of the cache.
var cacheCapacityChan = make(chan chan string)
// The channel where a bool passed into it will cause the OperateCache function to be closed and all of the data to be cleared.
var cacheCloseChan = make(chan bool)


var chancachent = make(chan * cacheEntry)

var checkcache = make(chan string)

//var stillprocessing = make(chan int)

// A wrapper function that does the actual getting of the file from the cache.
func getFile(filename string) (response *fileResponse) {
	// You need to add sanity checking here: The requested file
	// should be made relative (strip out leading "/" characters,
	// then have a "./" put on the start, and if there is ever the
	// string "/../", replace it with "/", the string "\/" should
	// be replaced with "/", and finally any instances of "//" (or
	// more) should be replaced by a single "/".
	// Hint: A replacement may lead to needing to do more replacements!

	// Also if you get a request which is just "/", you should return the file "./index.html"

	// You should also return a timeout error (take a look at the userlib) after `timeout`
	// seconds if there is no response from the disk.

	/*** YOUR CODE HERE ***/

	new := "." + filename
	for {
		if !(strings.Contains(new, "/../" ) || strings.Contains(new, "\\/" ) || strings.Contains(new, "//" )) {
			break;
		}
		new = strings.Replace(new, "/../", "/", -1)
		new = strings.Replace(new, "\\/", "/", -1)
		new = strings.Replace(new, "//", "/", -1)
	} 
	
	lastchar := new[len(new) - 1:len(new)]
	if lastchar == "/" {
		new = new + "index.html"
	}
	
	filename = new


	/*** YOUR CODE HERE END ***/

	// The part below will make a request on the fileChan and wait for a response to be issued from the cache.
	// You should not really need to modify anything below here.
	// Makes the file request object.
	request := fileRequest{filename, make(chan *fileResponse)}
	// Sends a pointer to the file request object to the fileChan so the cache can process the file request.
	fileChan <- &request
	// Returns the result (from the fileResponse channel)
	return <- request.response
}

// This function returns a string of the cache current status.
// It will just make a request to the cache asking for the status.
// You should not need to modify this function.
func getCacheStatus() (response string) {
	// Make a channel for the response of the Capacity request.
	responseChan := make(chan string)
	// Send the response channel to the capacity request channel.
	cacheCapacityChan <- responseChan
	// Return the reply.
	return <- responseChan
}

// This function will tell the cache that it needs to close itself.
// You should not need to modify this function.
func CacheClear() (response string) {
	// Send the response channel to the capacity request channel.
	cacheCloseChan <- true
	// We should only return to here once we are sure the currently open cache will not process any more requests.
	// This is because the close channel is blocking until it pulls the item out of there.
	// Now that the cache should be closed, lets relaunch the cache.
	go operateCache()
	return userlib.CacheCloseMessage
}


type cacheEntry struct {
	filename string
	data []byte

	// You may want to add other stuff here...
}


func runReadFile(chanreq * fileRequest, channelbyte2 chan * []byte, file string) {
	response, err := userlib.ReadFile(workingDir, file)
	println("Finished reading from file!")
	if err != nil {
		// If we have an error from the read, will return the generic file error message and set the error code to follow that.
		//println("About to call chanreq response")
		chanreq.response <- &fileResponse{file, nil, fmt.Errorf("FileError exists"), nil}
	} else {
		//println("About to call channelbyte response")
		channelbyte2 <- &response
	}
}

func completeReadFile(chanreq * fileRequest, filename string, channelent chan * cacheEntry) {

	//Fix this to only send fileresponse if not timeout error or regular error, but still want to update cache regardless. Now
	//always sending file response


	holdchannelbyte := make(chan * []byte)
	go runReadFile(chanreq, holdchannelbyte, filename)
	var databyte []byte
	println("Time out is:")
	println(timeout)
	//istimeout := 0
	select {
		case <- time.After(time.Duration(timeout) * time.Second):
			//istimeout = 1
			println("Timeout about to be sent")
			chanreq.response <- &fileResponse{filename, nil, fmt.Errorf("TimeOut Error exists"), nil}
			bytedata := <-holdchannelbyte
			databyte = *bytedata
			channelent <- &cacheEntry{filename, databyte}
			//println("Timeout about to be sent")
			//chanreq.response <- &fileResponse{filename, nil, fmt.Errorf("TimeOut Error exists"), nil}
		case bytedata := <-holdchannelbyte:
			println("Byte data received first")
			databyte = *bytedata
			channelent <- &cacheEntry{filename, databyte}
			chanreq.response <- &fileResponse{filename, databyte, nil, nil}
	}

	//if istimeout == 0 {
	//	chanreq.response <- &fileResponse{filename, databyte, nil, nil}
	//} else {
	//		chanreq.response <- &fileResponse{filename, nil, fmt.Errorf("TimeOut Error exists"), nil}
	//	}


	// select {
	// 	case <-checkcache:
	// 		if istimeout == 0 {
	// 			//println("About to pass into cachenetry")
	// 			chanreq.response <- &fileResponse{filename, databyte, nil, nil}
	// 		} else {
	// 			//println("Calling timeoutresponse now")
	// 			chanreq.response <- &fileResponse{filename, nil, fmt.Errorf("TimeOut Error exists"), nil}	
	// 		}
	// }
	// println("End of filechan!")
}

// This function is where you need to do all the work...
// Basically, you need to...

// 1:  Create a map to store all the cache entries.

// 2:  Go into a continual select loop.

// Hint, you are going to want another channel 



func operateCache() {
	/* TODO Initialize your cache and the service requests until the program exits or you receive a message on the
	 * cacheCloseChan at which point you should clean up (aka clear any caching global variables and return from
	 * this function. */
 	// HINT: Take a look at the global channels given above!
	/*** YOUR CODE HERE ***/
	// Make a file map (this is just like a hashmap in java) for the cache entries.

	filemap := make(map[string]cacheEntry)
	cachecap := 0
	//Maybe make it a channel?
 

	//Don't touch your map anywahere inside of a goroutine. Having multiple instances trying to access craeates a race. 


	// Pull out of FileChan
		// - Hint:Add another channel fileRead
		
		// - Check if it is in map. If it is in map: then you know what what to do
		// - If not in map:
		// 	- Then spin a goroutine, and get a file in readfile, and create the response, put stuff in the cache 
		// - Have 2 gofunctions, a go within a go
			//timeout should be within the inner switch statement

		//Don't use a goroutine for putting stuff in a cache
		//Make sure to preserve the file request, which should pass into "filechan" the same filerequest as before 

	// Once you have made a filemap, here is a good skeleton for oyu to use to handle requests.
	for {
		// We want to select what we want to do based on what is in different cache channels.
		select {
		case fileReq := <- fileChan:
			println("Entering filChan!")
			fileReq = fileReq
			value, incache := filemap[fileReq.filename]
    		if incache {
    			fileReq.response <- &fileResponse{value.filename, value.data, nil, fileReq.response}
    		} else {
    			go completeReadFile(fileReq, fileReq.filename, chancachent)
    		}
		
		case cachdiskent := <-chancachent:
			if cachdiskent != nil {
				cachdiskdata := cachdiskent.data
				cachdiskfilename := cachdiskent.filename
				cachevalue, alrin := filemap[cachdiskfilename]
				cachevalue = cachevalue
				if !alrin {
					lockey := ""
					if len(cachdiskdata) <= capacity {
						println("Adding to cache!")
						for len(cachdiskdata)+cachecap > capacity {
							for k := range filemap {
								lockey = k
								break
							}
							cachecap = cachecap - len(filemap[lockey].data)
							delete(filemap, lockey)
						}
						filemap[cachdiskfilename] = cacheEntry{cachdiskfilename, cachdiskdata}
						cachecap = len(cachdiskdata) + cachecap
					}
				}
			}
			//println("Complete cachdisk")
			//checkcache <- "Complete"

		case cacheReq := <- cacheCapacityChan:
			//println("CacheCapcityChan coming in now!")
			numberofent := len(filemap)
			capacityofcache := cachecap
			capacitymax := capacity
			cacheReq <- fmt.Sprintf(userlib.CapacityString, numberofent, capacityofcache, capacitymax)

		case <- cacheCloseChan:
			// We want to exit the cache.
			// Make sure you clean up all of your cache state or you will fail most all of the tests!
			//println("CachCloseChan coming in now!")
			for skey := range filemap {
				cachecap = cachecap - len(filemap[skey].data)
    			delete(filemap, skey)
			}

			//
			//Clear global variable that cache may have seet up over here
			//close(chancachent)
			//close(checkcache)
			return
		}
	}
	/*** YOUR CODE HERE END ***/
}


// This functions when you do `go run server.go`. It will read and parse the command line arguments, set the values
// of some global variables, print out the server settings, tell the `http` library which functions to call when there
// is a request made to certain paths, launch the cache, and finally listen for connections and serve the requests
// which a connection may make. When it services a request, it will call one of the handler functions depending on if
// the prefix of the path matches the pattern which was set by the HandleFunc.
// You should not need to modify any of this.
func main(){
	// Initialize the arguments when the main function is ran. This is to setup the settings needed by
	// other parts of the file server.
	flag.IntVar(&port, "p", 8080, "Port to listen for HTTP requests (default port 8080).")
	flag.IntVar(&capacity, "c", 100000, "Number of bytes to allow in the cache.")
	flag.IntVar(&timeout, "t", 2, "Default timeout (in seconds) to wait before returning an error.")
	flag.StringVar(&workingDir, "d", "public_html/", "The directory which the files are hosted in.")
	// Parse the args.
	flag.Parse()
	// Say that we are starting the server.
	fmt.Printf("Server starting, port: %v, cache size: %v, timout: %v, working dir: '%s'\n", port, capacity, timeout, workingDir)
	serverString := fmt.Sprintf(":%v", port)

	// Set up the service handles for certain pattern requests in the url.
	http.HandleFunc("/", handler)
	http.HandleFunc("/cache/", cacheHandler)
	http.HandleFunc("/cache/clear/", cacheClearHandler)

	// Start up the cache logic...
	go operateCache()

	// This starts the web server and will cause it to continue to listen and respond to web requests.
	log.Fatal(http.ListenAndServe(serverString, nil))
}
