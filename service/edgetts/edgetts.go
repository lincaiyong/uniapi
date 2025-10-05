package edgetts

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/lincaiyong/log"
	"html"
	"math"
	"net/http"
	"strings"
	"time"
)

func generateSecMsGec(edgeToken string) string {
	ticks := float64(time.Now().UTC().Unix() + 11644473600)
	ticks = (ticks - math.Mod(ticks, 300)) * 1e7
	strToHash := fmt.Sprintf("%.0f%s", ticks, edgeToken)
	hash := sha256.Sum256([]byte(strToHash))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func EdgeTTS(text string) ([]byte, error) {
	edgeToken := "6A5AA1D4EAFF4E9FB37E23D68491D6F4" // check https://github.com/rany2/edge-tts/blob/master/src/edge_tts/constants.py
	baseUrl := "api.msedgeservices.com/tts/cognitiveservices/websocket/v1"
	userAgent := "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36 Edg/140.0.0.0"
	connectId := strings.ReplaceAll(uuid.New().String(), "-", "")
	token := generateSecMsGec(edgeToken)
	wssUrl := fmt.Sprintf("wss://%s?Ocp-Apim-Subscription-Key=%s&ConnectionId=%s&Sec-MS-GEC=%s&Sec-MS-GEC-Version=1-140.0.3485.14",
		baseUrl, edgeToken, connectId, token)
	conn, _, err := websocket.DefaultDialer.Dial(wssUrl, http.Header{
		"User-Agent": []string{userAgent},
	})
	if err != nil {
		return nil, fmt.Errorf("fail to connect to websocket server: %w", err)
	}
	defer func() { _ = conn.Close() }()
	log.InfoLog("connected to websocket server")

	timestamp1 := time.Now().UTC().Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
	configMsg := strings.ReplaceAll(fmt.Sprintf(`X-Timestamp:%s
Content-Type:application/json; charset=utf-8
Path:speech.config

{
  "context": {
    "synthesis": {
      "audio":{
        "metadataoptions": {
          "sentenceBoundaryEnabled": "true",
          "wordBoundaryEnabled":"false"
        },
        "outputFormat": "riff-16khz-16bit-mono-pcm"
      }
    }
  }
}`, timestamp1), "\n", "\r\n")
	err = conn.WriteMessage(websocket.TextMessage, []byte(configMsg))
	if err != nil {
		return nil, fmt.Errorf("fail to send config message: %w", err)
	}
	log.InfoLog("sent config message")

	timestamp2 := time.Now().UTC().Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
	ssmlMsg := strings.ReplaceAll(fmt.Sprintf(`X-RequestId:f28cb19ed4244cf6931eb123ba09ce9c
Content-Type:application/ssml+xml
X-Timestamp:%s
Path:ssml

<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='en-US'>
  <voice name='Microsoft Server Speech Text to Speech Voice (en-US, EmmaMultilingualNeural)'>
    <prosody pitch='+0Hz' rate='+0%%' volume='+0%%'>
      %s
    </prosody>
  </voice>
</speak>`,
		timestamp2, html.EscapeString(text)), "\n", "\r\n")
	err = conn.WriteMessage(websocket.TextMessage, []byte(ssmlMsg))
	if err != nil {
		return nil, fmt.Errorf("fail to send ssml message: %w", err)
	}
	log.InfoLog("sent ssml message")
	var audioData []byte
	for {
		messageType, message, readErr := conn.ReadMessage()
		if readErr != nil {
			return nil, fmt.Errorf("fail to read message: %w", err)
		}
		if messageType == websocket.BinaryMessage {
			log.InfoLog("receive binary message: %d bytes", len(message))
			headerIndex := strings.Index(string(message), "Path:audio\r\n")
			if headerIndex > 0 {
				pureAudioData := message[headerIndex+len("Path:audio\r\n"):]
				audioData = append(audioData, pureAudioData...)
			} else {
				return nil, fmt.Errorf("fail to parse audio data")
			}
		} else {
			log.InfoLog("receive text message: %d bytes", len(message))
			if strings.Contains(string(message), "Path:turn.end") {
				if len(audioData) == 0 {
					return nil, fmt.Errorf("audio data is empty")
				}
				return audioData, nil
			}
		}
	}
}
