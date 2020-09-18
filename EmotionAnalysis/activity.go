package sample

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/project-flogo/core/activity"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
	"gocv.io/x/gocv"
	//"github.com/project-flogo/core/data/metadata"
	// 	"reflect"
)

var (
	activityMd               = activity.ToMetadata(&Input{}, &Output{})
	model                    *tf.SavedModel
	maxValueIndex            int
	age                      string
	textColor                = color.RGBA{0, 255, 0, 0}
	pt                       = image.Pt(20, 20)
	left, top, right, bottom int
	frameIndex               = 0
	width, height            = 0, 0
	lastDemoID               = -1
	demoID                   = 1

	// demoID, _ = strconv.Atoi(os.Getenv("DEMOID"))
	// ageStage [3]string
	// window = gocv.NewWindow("Age")
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

func init() {
	_ = activity.Register(&Activity{})
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
	// fmt.Printf("Input serial: %s\n", input.Serial)
	imgjsonS := input.Serial
	imgjsonS = strings.Replace(imgjsonS, "\\\"", "\"", -1)
	// fmt.Printf("\n %c[%d;%d;%dmInput serial: %s%c[0m\n", 0x1B, 0, 0, 31, imgjsonS, 0x1B)
	//recognition done here, dummy now

	imgjson := imgJson{}
	json.Unmarshal([]byte(imgjsonS), &imgjson)
	fmt.Println(imgjson)

	demoID = imgjson.DemoID
	if demoID > 0 {
		return true, nil
	}
	tStart := time.Now().UnixNano()
	fmt.Println("\nStart Time for EmotionAnalysis Activity: ", tStart)
	if demoID != lastDemoID {
		model, _ = tf.LoadSavedModel("resource/EmotionModel", []string{"serve"}, nil)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		lastDemoID = demoID
	}

	// This flow does not work when demoID is 1.

	// // *************************
	// // imgName := "tmpAge.jpg"
	// receiveString := input.Serial
	// faceArr := strings.Split(receiveString, ";")
	// framePath := faceArr[0]
	framePath := imgjson.Imgpath
	// var ages []string

	// // ***************************************
	if exists(framePath) {
		for faceIndex := 0; faceIndex < len(imgjson.Bboxes); faceIndex++ {
			img := gocv.IMRead(framePath, gocv.IMReadColor)
			if frameIndex == 0 {
				dims := img.Size()
				fmt.Println(dims)
				width, height = dims[1], dims[0]
			}
			left := imgjson.Bboxes[faceIndex].X1
			top := imgjson.Bboxes[faceIndex].Y1
			right := imgjson.Bboxes[faceIndex].X2
			bottom := imgjson.Bboxes[faceIndex].Y2
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
			gocv.IMWrite("resource/temp/tmpEmo.jpg", imgFace)
			imgName := "resource/temp/tmpEmo.jpg"

			labels := []string{
				"Angry", "Disgust", "Fear", "Happy", "Neutral",
				"Sad", "Surprise",
			}

			imageFile, err := os.Open(imgName)
			if err != nil {
				log.Fatal(err)
			}
			var imgBuffer bytes.Buffer
			io.Copy(&imgBuffer, imageFile)
			imgtf, err := readImage(&imgBuffer, "jpg")
			if err != nil {
				log.Fatal("error making tensor: ", err)
			}

			// plTensor, _ := tf.NewTensor(false)
			result, err := model.Session.Run(
				map[tf.Output]*tf.Tensor{
					model.Graph.Operation("conv2d_54_input").Output(0): imgtf,
					// model.Graph.Operation("Placeholder").Output(0): plTensor,
				},
				[]tf.Output{
					// model.Graph.Operation("Softmax").Output(0),
					model.Graph.Operation("activation_10/Softmax").Output(0),
				},
				nil,
			)

			if err != nil {
				log.Fatal(err)
			}

			if preds, ok := result[0].Value().([][]float32); ok {
				// fmt.Println(preds)
				maxValueIndex = indexOfMax(preds[0])
			}
			// fmt.Println(maxValueIndex)
			emr := labels[maxValueIndex]
			fmt.Printf("\n %c[%d;%d;%dmEmotion Analysis Result: %s%c[0m\n", 0x1B, 0, 0, 32, emr, 0x1B)
			// fmt.Println()
		}

		// imgjsonr := imgJsonR{
		// 	ImgJson: imgjson,
		// 	Result:  ages}
		// imgjsonrB, _ := json.Marshal(&imgjsonr)
		// imgjsonrS := string(imgjsonrB)
		// // fmt.Println(imgjsonrS)
		// fmt.Printf("\n %c[%d;%d;%dmResult: %s%c[0m\n", 0x1B, 0, 0, 32, imgjsonrS, 0x1B)
		// imgjsonrS = strings.Replace(imgjsonrS, "\"", "\\\"", -1) //Escape character
		output := &Output{DisplayJson: "Emotion Analysis Done!"}
		// output := &Output{Serial: `te[{:,"st`}
		err = ctx.SetOutputObject(output)
		if err != nil {
			return true, err
		}

		// *******************************
		ctx.Logger().Debugf("Input serial: %s", input.Serial)
		// 	ctx.Logger().Debugf("Age: %s", age)

	}

	tEnd := time.Now().UnixNano()
	fmt.Println("\nEnd Time for EmotionAnalysis Activity: ", tEnd)
	fmt.Println("\nThe Time Consumption for EmotionAnalysis Activity: ", tEnd-tStart)
	fmt.Println("\nFrame ", frameIndex, "done.")
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

func readImage(imageBuffer *bytes.Buffer, imageFormat string) (*tf.Tensor, error) {
	tensor, err := tf.NewTensor(imageBuffer.String())
	if err != nil {
		return nil, err
	}
	graph, input, output, err := transformGraph(imageFormat)
	if err != nil {
		return nil, err
	}
	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	normalized, err := session.Run(
		map[tf.Output]*tf.Tensor{input: tensor},
		[]tf.Output{output},
		nil)
	if err != nil {
		return nil, err
	}
	return normalized[0], nil
}

func transformGraph(imageFormat string) (graph *tf.Graph, input,
	output tf.Output, err error) {
	var H, W int32
	// if demoID < 3 {
	// 	H, W = 224, 224
	// } else {
	// 	H, W = 160, 160
	// }
	H, W = 48, 48

	const (
		// H, W  = 160, 160
		Mean  = float32(117)
		Scale = float32(1)
	)
	s := op.NewScope()
	input = op.Placeholder(s, tf.String)

	var decode tf.Output
	switch imageFormat {
	case "png":
		decode = op.DecodePng(s, input, op.DecodePngChannels(3))
	case "jpg",
		"jpeg":
		decode = op.DecodeJpeg(s, input, op.DecodeJpegChannels(1))
	default:
		return nil, tf.Output{}, tf.Output{},
			fmt.Errorf("imageFormat not supported: %s", imageFormat)
	}

	output = op.Div(s,
		op.Sub(s,
			op.ResizeBilinear(s,
				op.ExpandDims(s,
					op.Cast(s, decode, tf.Float),
					op.Const(s.SubScope("make_batch"), int32(0))),
				op.Const(s.SubScope("size"), []int32{H, W})),
			op.Const(s.SubScope("mean"), Mean)),
		op.Const(s.SubScope("scale"), Scale))
	graph, err = s.Finalize()
	return graph, input, output, err
}

func indexOfMax(arr []float32) int {

	//Get the maximum value in an array and get the index

	//Declare an array
	// var arr [5]int = [...]int{6, 45, 63, 16, 86}
	//Suppose the first element is the maximum value and the index is 0.
	maxVal := arr[0]
	maxIndex := 0

	for i := 1; i < len(arr); i++ {
		//Cycle comparisons from the second element, exchange if found to be larger
		if maxVal < arr[i] {
			maxVal = arr[i]
			maxIndex = i
		}
	}

	return maxIndex
}
