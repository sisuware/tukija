//author: https://github.com/5k3105

package main

import (
	"strconv"
  "encoding/json"
  "fmt"
  "log"
  "io/ioutil"
  "net/http"
  "net/url"
  "os"
  "os/user"
  "path/filepath"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/google"
  "google.golang.org/api/youtube/v3"
)

var (
	Scene     *widgets.QGraphicsScene
	View      *widgets.QGraphicsView
	Item      *widgets.QGraphicsPixmapItem
	statusbar *widgets.QStatusBar
	mp        bool
  service *youtube.Service
)

func messageBox(message string, error error) (*widgets.QMessageBox) {
  messageBox := widgets.NewQMessageBox(nil)
  messageBox.SetText(message)

  if error != nil {
    messageBox.SetDetailedText(error.Error())
  }

  return messageBox
}

const missingClientSecretsMessage = `
Please configure OAuth 2.0
`

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
  cacheFile, err := tokenCacheFile()

  if err != nil {
    msg := messageBox("Unable to get path to cached credential file. %v", err)
    msg.Show()
  }

  tok, err := tokenFromFile(cacheFile)

  if err != nil {
    tok = getTokenFromWeb(config)
    saveToken(cacheFile, tok)
  }

  return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
  authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
  msg := messageBox(authURL, nil)
  msg.Exec()

  

  // if  msg.Result() {
  //   fmt.Printf('ok')
  // }

  var code string
  inputBox := widgets.NewQInputDialog(nil, 0)
  inputBox.SetLabelText("Code")

  if _, err := fmt.Scan(&code); err != nil {
    msg := messageBox("Unable to read authorization code %v", err)
    msg.Show()
  }

  tok, err := config.Exchange(oauth2.NoContext, code)

  if err != nil {
    msg := messageBox("Unable to retrieve token from web %v", err)
    msg.Show()
  }

  return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
  usr, err := user.Current()
  if err != nil {
    return "", err
  }
  tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
  os.MkdirAll(tokenCacheDir, 0700)
  return filepath.Join(tokenCacheDir,
    url.QueryEscape("tukija.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
  f, err := os.Open(file)
  if err != nil {
    return nil, err
  }
  t := &oauth2.Token{}
  err = json.NewDecoder(f).Decode(t)
  defer f.Close()
  return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
  fmt.Printf("Saving credential file to: %s\n", file)

  f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

  if err != nil {
    log.Fatalf("Unable to cache oauth token: %v", err)
  }

  defer f.Close()
  json.NewEncoder(f).Encode(token)
}

func auth() {
  ctx := context.Background()

  b, err := ioutil.ReadFile("client_secret.json")

  if err != nil {
    msg := messageBox("Unable to read client secret file: %v", err)
    msg.Show()
  }

  // If modifying these scopes, delete your previously saved credentials
  // at ~/.credentials/youtube-go-quickstart.json
  config, err := google.ConfigFromJSON(b, youtube.YoutubeChannelMembershipsCreatorScope)

  if err != nil {
    msg := messageBox("Unable to parse client secret file to config: %v", err)
    msg.Show()
  }

  client := getClient(ctx, config)
  service, err = youtube.New(client)

  msg := messageBox("Error creating YouTube client: %v", err)
  msg.Show()
}

func main() {
	widgets.NewQApplication(len(os.Args), os.Args)

	// Main Window
	var window = widgets.NewQMainWindow(nil, 0)
	window.SetWindowTitle("Tukija")
	window.SetMinimumSize2(360, 520)

  // About Menu
  aboutMenu := widgets.NewQMenu(nil)
  aboutMenu.SetTitle("About")

  menuBar := widgets.NewQMenuBar(nil)
  menuBar.AddMenu(aboutMenu)
  window.SetMenuBar(menuBar)


	// Statusbar
	statusbar = widgets.NewQStatusBar(window)
	window.SetStatusBar(statusbar)

	Scene = widgets.NewQGraphicsScene(nil)
	View = widgets.NewQGraphicsView(nil)

	Scene.ConnectKeyPressEvent(keyPressEvent)
	Scene.ConnectWheelEvent(wheelEvent)
	View.ConnectResizeEvent(resizeEvent)

	// dx, dy := 16, 32
  //
	// img := gui.NewQImage3(dx, dy, gui.QImage__Format_ARGB32)
  //
	// for i := 0; i < dx; i++ {
	// 	for j := 0; j < dy; j++ {
	// 		img.SetPixelColor2(i, j, gui.NewQColor3(i*2, j*8, i*2, 255))
  //
	// 	}
	// }
  //
	// //img = img.Scaled2(dx*2,dy,core.Qt__IgnoreAspectRatio, core.Qt__FastTransformation)
  //
	// Item = widgets.NewQGraphicsPixmapItem2(gui.NewQPixmap().FromImage(img, 0), nil)
  //
	// Item.ConnectMouseMoveEvent(ItemMouseMoveEvent)
	// Item.ConnectMousePressEvent(ItemMousePressEvent)
	// Item.ConnectMouseReleaseEvent(ItemMouseReleaseEvent)
  //
	// Item.SetAcceptHoverEvents(true)
	// Item.ConnectHoverMoveEvent(ItemHoverMoveEvent)
  //
	// Scene.AddItem(Item)
  //
	// View.SetScene(Scene)
	// View.Show()
  //
	// statusbar.ShowMessage(core.QCoreApplication_ApplicationDirPath(), 0)
  //
	// // Set Central Widget
	// window.SetCentralWidget(View)

	// Run App
	// widgets.QApplication_SetStyle2("fusion")
	window.Show()

  auth()

	widgets.QApplication_Exec()
}

func ItemMousePressEvent(event *widgets.QGraphicsSceneMouseEvent) {
	mp = true
	mousePosition := event.Pos()
	x, y := int(mousePosition.X()), int(mousePosition.Y())
	drawpixel(x, y)

}

func ItemMouseReleaseEvent(event *widgets.QGraphicsSceneMouseEvent) {
	mp = false

	Item.MousePressEventDefault(event) // absofukinlutely necessary for drag & draw !!

	//Item.MouseReleaseEventDefault(event) // worthless
}

func ItemMouseMoveEvent(event *widgets.QGraphicsSceneMouseEvent) {
	mousePosition := event.Pos()
	x, y := int(mousePosition.X()), int(mousePosition.Y())

	drawpixel(x, y)

}

func ItemHoverMoveEvent(event *widgets.QGraphicsSceneHoverEvent) {
	mousePosition := event.Pos()
	x, y := int(mousePosition.X()), int(mousePosition.Y())

	rgbValue := Item.Pixmap().ToImage().PixelColor2(x, y)
	r, g, b := rgbValue.Red(), rgbValue.Green(), rgbValue.Blue()
	statusbar.ShowMessage("x: "+strconv.Itoa(x)+" y: "+strconv.Itoa(y)+" r: "+strconv.Itoa(r)+" g: "+strconv.Itoa(g)+" b: "+strconv.Itoa(b), 0)

}

func drawpixel(x, y int) {

	if mp {
		img := Item.Pixmap().ToImage()
		img.SetPixelColor2(x, y, gui.NewQColor3(255, 255, 255, 255))
		Item.SetPixmap(gui.NewQPixmap().FromImage(img, 0))
	}

}

func keyPressEvent(e *gui.QKeyEvent) {

	switch int32(e.Key()) {
	case int32(core.Qt__Key_0):
		View.Scale(1.25, 1.25)

	case int32(core.Qt__Key_9):
		View.Scale(0.8, 0.8)
	}

}

func wheelEvent(e *widgets.QGraphicsSceneWheelEvent) {
	if gui.QGuiApplication_QueryKeyboardModifiers()&core.Qt__ShiftModifier != 0 {
		if e.Delta() > 0 {
			View.Scale(1.25, 1.25)
		} else {
			View.Scale(0.8, 0.8)
		}
	}
}

func resizeEvent(e *gui.QResizeEvent) {

	View.FitInView(Scene.ItemsBoundingRect(), core.Qt__KeepAspectRatio)

}
