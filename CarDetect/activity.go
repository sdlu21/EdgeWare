package sample

import (
	"fmt"
	"image"
	"strconv"
	"time"

	"github.com/project-flogo/core/activity"
	"gocv.io/x/gocv"

	//	"github.com/project-flogo/core/data/metadata"
	// "image/color"
	// "image"
	// "log"

	"bufio"
	"bytes"

	// "fmt"
	// "image"
	"io/ioutil"
	"log"
	"os"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/text"

	// "gocv.io/x/gocv"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"

	_ "golang.org/x/image/bmp"
)

var (
	activityMd = activity.ToMetadata(&Input{}, &Output{})
	// window     = gocv.NewWindow("EdgeWare")
	img       gocv.Mat
	blob      gocv.Mat
	err       error
	setDemoID = 1
	// deviceID  = "resource/645c610.avi"
	webcam, _ = gocv.OpenVideoCapture("resource/645c610.avi")
	// net        = gocv.ReadNet("resource/tfCarModel/frozen_inference_graph.pb", "resource/tfCarModel/ssd_mobilenet_v1_coco_2017_11_17.pbtxt")

	// default name for frozen graph file
	// COCO labels file

	// Capture frame width
	W = 640
	// Capture frame height
	H = 480

	// global TF session, re-usable and concurrency safe
	session *tf.Session
	// global model graph
	graph *tf.Graph
	// global slice of labels
	labels = loadLabels("resource/carModel/data/coco_labels.txt")
	// Webcam device ID for OpenCV

	model, _   = ioutil.ReadFile("resource/carModel/models/ssd_mobilenet_v1_coco_11_06_2017/frozen_inference_graph.pb")
	lastDemoID = 1

	// // Setup Pixel window
	// cfg = pixelgl.WindowConfig{
	// 	Title:  "EdgeWare",
	// 	Bounds: pixel.R(0, 0, 640, 480),
	// 	VSync:  true,
	// }
	// win, _ = pixelgl.NewWindow(cfg)

	// // Setup Pixel requirements for drawing boxes and labels
	// mat = pixel.IM.Moved(win.Bounds().Center())

	atlas = text.NewAtlas(basicfont.Face7x13, text.ASCII)
	imd   = imdraw.New(nil)
)

func init() {
	_ = activity.Register(&Activity{}) //activity.Register(&Activity{}, New) to create instances using factory method 'New'

}

//New optional factory method, should be used if one activity instance per configuration is desired
func New(ctx activity.InitContext) (activity.Activity, error) {

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

// Eval implements api.Activity.Eval - Logs the Message
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {

	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return true, err
	}
	tStart := time.Now().UnixNano()
	fmt.Println("\nStart Time for CarDetect Activity: ", tStart)
	fmt.Printf("Input DemoId: %s\n", input.DemoId)

	if len(input.DemoId) > 0 {
		setDemoID, _ = strconv.Atoi(input.DemoId)
	}
	fmt.Printf("Set DemoId: %d\n", setDemoID)

	if setDemoID <= 3 {
		return true, nil
	}

	if setDemoID == 4 && setDemoID != lastDemoID {
		graph = tf.NewGraph()
		if err := graph.Import(model, ""); err != nil {
			log.Fatal(err)
		}

		// Create a session for inference over graph.
		session, err = tf.NewSession(graph, nil)
		if err != nil {
			log.Fatal(err)
		}

		// defer session.Close()

		lastDemoID = setDemoID
	}

	img = gocv.NewMat()
	defer img.Close()

	if ok := webcam.Read(&img); !ok {
		fmt.Printf("Device closed\n")
		return true, nil
	}

	gocv.Resize(img, &img, image.Point{X: W, Y: H}, 0.0, 0.0, gocv.InterpolationNearestNeighbor)

	// Encode Mat as a bmp (uncompressed)
	frame, err := gocv.IMEncode(".bmp", img)
	if err != nil {
		log.Fatalf("Error encoding frame: %v", err)
	}

	// for !win.Closed() {
	// Make a tensor and an Image from the frame bytes
	tensor, img1, err := makeTensorFromImage(frame)
	if err != nil {
		log.Fatalf("error making input tensor: %v", err)
	}

	// Run inference on the newly made input tensor
	probabilities, classes, boxes, err := predictObjectBoxes(tensor)
	if err != nil {
		log.Fatalf("error making prediction: %v", err)
	}

	// Turn our video frame into a a sprite to be drawn by Pixel
	pic := pixel.PictureDataFromImage(img1)
	// sprite := pixel.NewSprite(pic, pic.Bounds())

	// Clear any previous boxes
	imd.Clear()
	// // Clear previous spires
	// win.Clear(colornames.Black)
	// // Draw the new frame first
	// sprite.Draw(win, mat)

	// Draw a box around the objects
	curObj := 0
	// arbitrary detection threshold of 0.4
	for probabilities[curObj] > 0.4 {
		if int(classes[curObj]) > 14 {
			continue
		} //Only show traffic-related
		// box coordinates come in as [y1,x1,y2,x2]
		x1 := pic.Bounds().Max.X * float64(boxes[curObj][1])
		x2 := pic.Bounds().Max.X * float64(boxes[curObj][3])
		// TF (0,0) is the upper left, Pixel (0,0) is the lower left, so we need
		// to subtract the Y values from the max height so we draw from the bottom up
		y1 := pic.Bounds().Max.Y - (pic.Bounds().Max.Y * float64(boxes[curObj][0]))
		y2 := pic.Bounds().Max.Y - (pic.Bounds().Max.Y * float64(boxes[curObj][2]))

		// objColor := colornames.Map[colornames.Names[int(classes[curObj])]]
		objColor := colornames.Map["red"]

		// Draw the box label
		txt := text.New(pixel.V(x1, y1), atlas)
		txt.Color = objColor
		txt.WriteString(getLabel(curObj, probabilities, classes))
		// txt.Draw(win, pixel.IM.Scaled(txt.Orig, 2))

		// Push the box onto the draw stack
		imd.Color = objColor
		imd.Push(pixel.V(x1, y1), pixel.V(x2, y2))
		imd.Rectangle(2.0)

		fmt.Printf("\n %c[%d;%d;%dm%s: %f, %f, %f, %f%c[0m\n", 0x1B, 0, 0, 31, labels[int(classes[curObj])], y1, x1, y2, x2, 0x1B)

		curObj++
	}

	// draw whatever's in the stack
	// imd.Draw(win)
	// win.Update()

	// }

	// pixelgl.Run(run)

	// window.IMShow(img)
	// window.WaitKey(1)

	// output := &Output{Serial: sendString} //should be serial of the record in the database
	output := &Output{Serial: "Car Detect Finished!"}
	// output := &Output{Serial: `te[{:,"st`}
	err = ctx.SetOutputObject(output)
	if err != nil {
		return true, err
	}

	tEnd := time.Now().UnixNano()
	fmt.Println("\nEnd Time for CarDetect Activity: %d", tEnd)
	fmt.Println("\nThe Time Consumption for CarDetect Activity: ", tEnd-tStart)
	// fmt.Println(fmt.Sprintf("\nFrame %d done.", frameIndex))
	return true, nil
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// func main() {

// 	// Construct an in-memory graph from the serialized form.

// 	pixelgl.Run(run)
// }

func getLabel(idx int, probabilities []float32, classes []float32) string {
	index := int(classes[idx])
	// label := fmt.Sprintf("%s (%2.0f%%)", labels[index], probabilities[idx]*100.0)
	label := fmt.Sprintf("%s", labels[index])

	return label
}

func loadLabels(labelsFile string) []string {
	file, err := os.Open(labelsFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	var mlabels []string
	for scanner.Scan() {
		mlabels = append(mlabels, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("ERROR: failed to read %s: %v", labelsFile, err)
	}
	return mlabels
}

// Read frames from the webcam in a goroutine to absorb some of the read latency.
// This helps gain ~2 FPS.
func capture(deviceID string, frames chan []byte) {
	cam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		log.Fatal("failed reading cam")
	}
	defer cam.Close()

	frame := gocv.NewMat()
	defer frame.Close()

	for {
		if ok := cam.Read(&frame); !ok {
			log.Fatal("failed reading cam")
		}

		// Resize the Mat in place.
		gocv.Resize(frame, &frame, image.Point{X: W, Y: H}, 0.0, 0.0, gocv.InterpolationNearestNeighbor)

		// Encode Mat as a bmp (uncompressed)
		buf, err := gocv.IMEncode(".bmp", frame)
		if err != nil {
			log.Fatalf("Error encoding frame: %v", err)
		}

		// Push the frame to the channel
		frames <- buf
	}
}

func run() {

	// Start up the background capture
	// framesChan := make(chan []byte)
	// go capture(deviceID, framesChan)

}

// Build a graph to decode bitmap input into the proper tensor shape
// The object detection models take an input of [1,?,?,3]
func decodeBitmapGraph() (g *tf.Graph, input, output tf.Output, err error) {
	s := op.NewScope()
	input = op.Placeholder(s, tf.String)
	output = op.ExpandDims(s,
		op.DecodeBmp(s, input, op.DecodeBmpChannels(3)),
		op.Const(s.SubScope("make_batch"), int32(0)))
	g, err = s.Finalize()
	return
}

// Make a tensor from jpg image bytes
func makeTensorFromImage(img []byte) (*tf.Tensor, image.Image, error) {

	// DecodeJpeg uses a scalar String-valued tensor as input.
	tensor, err := tf.NewTensor(string(img))
	if err != nil {
		return nil, nil, err
	}
	// Creates a tensorflow graph to decode the jpeg image
	g, input, output, err := decodeBitmapGraph()
	if err != nil {
		return nil, nil, err
	}
	// Execute that graph to decode this one image
	sess, err := tf.NewSession(g, nil)
	if err != nil {
		return nil, nil, err
	}
	defer sess.Close()
	normalized, err := sess.Run(
		map[tf.Output]*tf.Tensor{input: tensor},
		[]tf.Output{output},
		nil)
	if err != nil {
		return nil, nil, err
	}

	r := bytes.NewReader(img)
	i, _, err := image.Decode(r)
	if err != nil {
		return nil, nil, err
	}
	return normalized[0], i, nil
}

// Run the image through the model
func predictObjectBoxes(input *tf.Tensor) (probabilities, classes []float32, boxes [][]float32, err error) {
	// Get all the input and output operations
	inputop := graph.Operation("image_tensor")
	// Output ops
	o1 := graph.Operation("detection_boxes")
	o2 := graph.Operation("detection_scores")
	o3 := graph.Operation("detection_classes")
	o4 := graph.Operation("num_detections")

	output, err := session.Run(
		map[tf.Output]*tf.Tensor{
			inputop.Output(0): input,
		},
		[]tf.Output{
			o1.Output(0),
			o2.Output(0),
			o3.Output(0),
			o4.Output(0),
		},
		nil)
	if err != nil {
		log.Fatalf("Error running session: %v", err)
	}

	probabilities = output[1].Value().([][]float32)[0]
	classes = output[2].Value().([][]float32)[0]
	boxes = output[0].Value().([][][]float32)[0]

	return
}
