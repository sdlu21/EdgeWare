package sample

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"

	"github.com/project-flogo/core/activity"
	"gocv.io/x/gocv"
	//	"github.com/project-flogo/core/data/metadata"
	// 	"reflect"
)

var (
	activityMd               = activity.ToMetadata(&Input{})
	textColor                = color.RGBA{0, 255, 0, 0}
	pt                       = image.Pt(20, 20)
	left, top, right, bottom int
	frameIndex               = 0
	width, height            = 0, 0
	windowDispaly            = gocv.NewWindow("EdgeWare - Dispaly")
	lastDemoID               = -1
	demoID                   = 1

	// demoID, _ = strconv.Atoi(os.Getenv("DEMOID"))
	// gender string
	// window = gocv.NewWindow("Gender")
	// textColor = color.RGBA{0, 255, 0, 0}
	// pt = image.Pt(20, 20)
	// left, top, right, bottom int
)

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

type imgJsonR struct {
	ImgJson imgJson  `json:"imgjson"`
	Result  []string `json:"result"`
}

var (
	lastImgJsonR = imgJsonR{}
	curImgJsonR  = imgJsonR{}
	ifDisplay    = false
)

func init() {
	_ = activity.Register(&Activity{}) //activity.Register(&Activity{}, New) to create instances using factory method 'New'
}

//New optional factory method, should be used if one activity instance per configuration is desired
func New(ctx activity.InitContext) (activity.Activity, error) {

	act := &Activity{} //add aSetting to instance

	return act, nil
}

// Activity is an sample Activity that can be used as a base to create a custom activity
type Activity struct {
}

// Metadata returns the activity's metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements api.Activity.Eval - Logs the Message
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return true, err
	}
	//recognition done here, dummy now
	// *************************
	// fmt.Printf("Input serial: %s\n", input.Serial)
	imgjsonS := input.DisplayJson
	imgjsonS = strings.Replace(imgjsonS, "\\\"", "\"", -1)
	// fmt.Printf("\n %c[%d;%d;%dmInput serial: %s%c[0m\n", 0x1B, 0, 0, 31, imgjsonS, 0x1B)
	// imgName := "tmpAge.jpg"
	// imgName := input.Serial
	fmt.Printf("\n %c[%d;%d;%dmResult: %s%c[0m\n", 0x1B, 0, 0, 33, imgjsonS, 0x1B)

	imgjson := imgJsonR{}
	json.Unmarshal([]byte(imgjsonS), &imgjson)
	curImgJsonR = imgjson
	// fmt.Println(imgjson)

	demoID = imgjson.ImgJson.DemoID
	if demoID != lastDemoID {
		lastDemoID = demoID
	}
	if demoID > 1 {
		if curImgJsonR.ImgJson.Imgid == lastImgJsonR.ImgJson.Imgid {
			ifDisplay = true
		} else {
			ifDisplay = false
		}

		if ifDisplay {
			framePath := curImgJsonR.ImgJson.Imgpath
			if exists(framePath) {
				for faceIndex := 0; faceIndex < len(curImgJsonR.ImgJson.Bboxes); faceIndex++ {
					img := gocv.IMRead(framePath, gocv.IMReadColor)
					if frameIndex == 0 {
						dims := img.Size()
						fmt.Println(dims)
						width, height = dims[1], dims[0]
					}
					left := curImgJsonR.ImgJson.Bboxes[faceIndex].X1
					top := curImgJsonR.ImgJson.Bboxes[faceIndex].Y1
					right := curImgJsonR.ImgJson.Bboxes[faceIndex].X2
					bottom := curImgJsonR.ImgJson.Bboxes[faceIndex].Y2
					left -= 20
					top -= 60
					right += 20
					bottom += 20
					if left < 0 {
						left = 0
					}
					if top < 0 {
						top = 0
					}
					if right > width {
						right = width
					}
					if bottom > height {
						bottom = height
					}
					rect := image.Rect(left, top, right, bottom)
					imgFace := img.Region(rect)
					age, gender := curImgJsonR.Result[faceIndex], lastImgJsonR.Result[faceIndex]
					if age == "male" || age == "female" {
						age, gender = gender, age
					}
					gocv.PutText(&imgFace, gender+", "+age, pt, gocv.FontHersheyPlain, 1.2, textColor, 2)
					windowDispaly.IMShow(imgFace)
					windowDispaly.WaitKey(30)
				}
				frameIndex++
			}
		}

		// *******************************
		lastImgJsonR = curImgJsonR

	} else {
		ifDisplay = true

		if ifDisplay {
			framePath := curImgJsonR.ImgJson.Imgpath
			if exists(framePath) {
				for faceIndex := 0; faceIndex < len(curImgJsonR.ImgJson.Bboxes); faceIndex++ {
					img := gocv.IMRead(framePath, gocv.IMReadColor)
					if frameIndex == 0 {
						dims := img.Size()
						fmt.Println(dims)
						width, height = dims[1], dims[0]
					}
					left := curImgJsonR.ImgJson.Bboxes[faceIndex].X1
					top := curImgJsonR.ImgJson.Bboxes[faceIndex].Y1
					right := curImgJsonR.ImgJson.Bboxes[faceIndex].X2
					bottom := curImgJsonR.ImgJson.Bboxes[faceIndex].Y2
					left -= 20
					top -= 60
					right += 20
					bottom += 20
					if left < 0 {
						left = 0
					}
					if top < 0 {
						top = 0
					}
					if right > width {
						right = width
					}
					if bottom > height {
						bottom = height
					}
					rect := image.Rect(left, top, right, bottom)
					imgFace := img.Region(rect)
					gender := curImgJsonR.Result[faceIndex]
					gocv.PutText(&imgFace, gender, pt, gocv.FontHersheyPlain, 1.2, textColor, 2)
					windowDispaly.IMShow(imgFace)
					windowDispaly.WaitKey(30)
				}
				frameIndex++
			}
		}

		// *******************************
		// lastImgJsonR = curImgJsonR

	}
	ctx.Logger().Debugf("Input serial: %s", input.DisplayJson)
	// 	ctx.Logger().Debugf("Age: %s", age)
	return true, nil

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// determine if the file/folder of the given path exists
func exists(path string) bool {

	_, err := os.Stat(path)
	//os.Stat get the file information
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
