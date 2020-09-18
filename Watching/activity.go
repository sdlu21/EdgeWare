package sample

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"time"

	"github.com/Kagami/go-face"
	"github.com/project-flogo/core/activity"
	"gocv.io/x/gocv"
	//	"github.com/project-flogo/core/data/metadata"
	// "image/color"
	// "image"
	// "log"
)

// const dataDir = "resource/faceModel"

var (
	activityMd = activity.ToMetadata(&Input{}, &Output{})
	window     = gocv.NewWindow("EdgeWare")
	img        gocv.Mat
	// rec        *face.Recognizer
	rec, _     = face.NewRecognizer("resource/faceModel")
	frameIndex = 0
	filename   string
	err        error
	imgDir     = os.Getenv("HOME") + "/flogo"
	setDemoID  = 1
	deviceID   = os.Getenv("DEVICEID")
	webcam, _  = gocv.OpenVideoCapture(deviceID)

	// DEVICEID can be 0 or any video file path
	//webcam, _ = gocv.OpenVideoCapture("resource/the_car_lab.mp4")
	//webcam, _ = gocv.OpenVideoCapture(0)

	//deviceID string
	//boxcolor color.RGBA
	//rec, _ = face.NewRecognizer("testdata")
)

func init() {
	_ = activity.Register(&Activity{}) //activity.Register(&Activity{}, New) to create instances using factory method 'New'
	// window = gocv.NewWindow("Flogo")
	// defer window.Close()
	// frameIndex = 0
	// img = gocv.NewMat()
	// defer img.Close()

	// // Init the recognizer.
	// rec, err = face.NewRecognizer(dataDir)
	// if err != nil {
	// 	log.Fatalf("Can't init face recognizer: %v", err)
	// }
	// // Free the resources when you're finished.
	// defer rec.Close()

	//*****************************************
	// deviceID = "the_car_lab.mp4"
	// // open capture device
	// webcam, err = gocv.OpenVideoCapture(deviceID)
	// if err != nil {
	// 	fmt.Printf("Error opening video capture device: %v\n", deviceID)
	// 	return
	// }

	// defer webcam.Close()

	// boxcolor = color.RGBA{0, 255, 0, 0}
}

//New optional factory method, should be used if one activity instance per configuration is desired
func New(ctx activity.InitContext) (activity.Activity, error) {

	//	s := &Settings{}
	//	err := metadata.MapToStruct(ctx.Settings(), s, true)
	//	if err != nil {
	//		return nil, err
	//	}

	//	ctx.Logger().Debugf("Setting: %s", s.ASetting)

	act := &Activity{} //add aSetting to instance//nothing to add now

	return act, nil
}

// Activity is an sample Activity that can be used as a base to create a custom activity
type Activity struct {
}

// Metadata returns the activity's metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// https://gobyexample.com/json
//bounding box by form of x1,y1,x2,y2
type Bbox struct {
	Boxid int `json:"boxid"`
	X1    int `json:"x1"`
	Y1    int `json:"y1"`
	X2    int `json:"x2"`
	Y2    int `json:"y2"`
}

//json format of person recognition
type imgJson struct {
	Imgid   int    `json:"imgid"`
	Imgpath string `json:"imgpath"`
	Bboxes  []Bbox `json:"bboxes"`
	DemoID  int    `json:"demoid"`
}

// Eval implements api.Activity.Eval - Logs the Message
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	tStart := time.Now().UnixNano()
	fmt.Println("\nStart Time for Watching Activity: ", tStart)
	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return true, err
	}
	fmt.Printf("Input DemoId: %s\n", input.DemoId)

	if len(input.DemoId) > 0 {
		setDemoID, _ = strconv.Atoi(input.DemoId)
	}
	fmt.Printf("Set DemoId: %d\n", setDemoID)

	img = gocv.NewMat()
	defer img.Close()

	if setDemoID > 3 {
		return true, nil
	}

	// deviceID = "the_car_lab.mp4"
	// open capture device
	// webcam, err = gocv.OpenVideoCapture(deviceID)
	// if err != nil {
	// 	fmt.Printf("Error opening video capture device: %v\n", deviceID)
	// 	return
	// }
	// defer webcam.Close()

	//call neural network here
	ctx.Logger().Debugf("result of picking out a person: %s", "found") //log is also dummy here
	err = nil                                                          //set if neural network go wrong
	if err != nil {
		return true, err
	}

	if deviceID == "0" {
		// If open webcamera, use the following code
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed: %v\n", deviceID)
			return
		}
		// fmt.Printf("read webcamera")
	} else {

		// If open mp4 file, use the following code
		for a := 0; a < 30; a++ {
			if ok := webcam.Read(&img); !ok {
				fmt.Printf("Device closed: %v\n", deviceID)
				return
			}
		}
	}
	// fmt.Println(img.Size())
	frameIndex++
	filename = imgDir + "/flogo" + strconv.Itoa(frameIndex) + ".jpg"
	fmt.Printf("\n %c[%d;%d;%dm %s%c[0m\n", 0x1B, 0, 0, 33, filename, 0x1B)

	window.IMShow(img)
	window.WaitKey(1)

	testImagePristin := "resource/temp/tmp.jpg"
	gocv.IMWrite(testImagePristin, img)

	// Recognize faces on that image.
	faces, err := rec.RecognizeFile(testImagePristin)
	if err != nil {
		log.Fatalf("Can't recognize: %v", err)
	}

	fmt.Printf("\n %c[%d;%d;%dm# of faces: %d%c[0m\n", 0x1B, 0, 0, 33, len(faces), 0x1B)
	// imgFace := gocv.IMRead(testImagePristin, gocv.IMReadColor)

	save := false
	// if save is true, indicating that the face is detected

	// sendString := filename
	var boxes []Bbox
	boxid := -1
	for _, f := range faces {
		mRect := f.Rectangle
		fmt.Println(mRect)
		// 	mRect.Min.X -= 20
		// 	mRect.Min.Y -= 60
		// 	mRect.Max.X += 20
		// 	mRect.Max.Y += 20
		// 	fmt.Println(mRect.Min.X, mRect.Min.Y, mRect.Max.X, mRect.Max.Y)
		// 	// gocv.Rectangle(&img, mRect, color.RGBA{0, 255, 0, 0}, 2)
		save = true
		// 	// rect := image.Rect(mRect.Min.X, mRect.Min.Y, mRect.Max.X, mRect.Max.Y)
		// 	// imgFace := img.Region(rect)

		// // 	frameIndex++
		// // 	filename = "/home/yyt/flogo/flogo" + strconv.Itoa(frameIndex) + ".jpg"
		// // 	gocv.IMWrite(filename, imgFace)
		// sendString += ";" + mRect.String()

		boxid++
		left := mRect.Min.X
		top := mRect.Min.Y
		right := mRect.Max.X
		bottom := mRect.Max.Y
		boxes = append(boxes, Bbox{Boxid: boxid, X1: left, Y1: top, X2: right, Y2: bottom})
	}

	//dummy json generation here
	//Imgid is at least 1
	imgjson := imgJson{
		Imgid:   frameIndex,
		Imgpath: filename,
		Bboxes:  boxes,
		DemoID:  setDemoID}
	imgjsonB, _ := json.Marshal(&imgjson)
	imgjsonS := string(imgjsonB)
	fmt.Println(imgjsonS)
	imgjsonS = strings.Replace(imgjsonS, "\"", "\\\"", -1) //Escape character

	// *************************

	// if !save {
	// 	return false, nil
	// }
	// ***********************
	// filename = testImagePristin
	//todo:
	// A frame of pictures may contain multiple faces, which will be stored as multiple files.
	// These file paths should be merged and transmitted in strings.
	// Now each picture only transmitted a face's path for testing

	//
	if save {
		gocv.IMWrite(filename, img)
		// output := &Output{Serial: sendString} //should be serial of the record in the database
		output := &Output{Serial: imgjsonS}
		// output := &Output{Serial: `te[{:,"st`}
		err = ctx.SetOutputObject(output)
		if err != nil {
			return true, err
		}
	}

	tEnd := time.Now().UnixNano()
	fmt.Println("\nEnd Time for Watching Activity: ", tEnd)
	fmt.Println("\nThe Time Consumption for Watching Activity: ", tEnd-tStart)
	fmt.Println("\nFrame ", frameIndex, "done.")
	return true, nil
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
