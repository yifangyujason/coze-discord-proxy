package controller

import (
	"coze-discord-proxy/common"
	"coze-discord-proxy/discord"
	"coze-discord-proxy/model"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Chat 发送消息
// @Summary 发送消息
// @Description 发送消息
// @Tags chat
// @Accept json
// @Produce json
// @Param chatModel body model.ChatReq true "chatModel"
// @Param proxy-secret header string false "proxy-secret"
// @Param out-time header string false "out-time"
// @Success 200 {object} model.ReplyResp "Successful response"
// @Router /api/chat [post]
//func Chat(c *gin.Context) {
//
//	var chatModel model.ChatReq
//	err := json.NewDecoder(c.Request.Body).Decode(&chatModel)
//	if err != nil {
//		common.LogError(c.Request.Context(), err.Error())
//		c.JSON(http.StatusOK, gin.H{
//			"message": "无效的参数",
//			"success": false,
//		})
//		return
//	}
//
//	sendChannelId, calledCozeBotId, err := getSendChannelIdAndCozeBotId(c, false, chatModel)
//	if err != nil {
//		common.LogError(c.Request.Context(), err.Error())
//		c.JSON(http.StatusOK, model.OpenAIErrorResponse{
//			OpenAIError: model.OpenAIError{
//				Message: "配置异常",
//				Type:    "invalid_request_error",
//				Code:    "discord_request_err",
//			},
//		})
//		return
//	}
//
//	sentMsg, err := discord.SendMessage(c, sendChannelId, calledCozeBotId, chatModel.Content)
//	if err != nil {
//		c.JSON(http.StatusOK, gin.H{
//			"success": false,
//			"message": err.Error(),
//		})
//		return
//	}
//
//	replyChan := make(chan model.ReplyResp)
//	discord.RepliesChans[sentMsg.ID] = replyChan
//	defer delete(discord.RepliesChans, sentMsg.ID)
//
//	stopChan := make(chan model.ChannelStopChan)
//	discord.ReplyStopChans[sentMsg.ID] = stopChan
//	defer delete(discord.ReplyStopChans, sentMsg.ID)
//
//	timer, err := setTimerWithHeader(c, chatModel.Stream, common.RequestOutTimeDuration)
//	if err != nil {
//		common.LogError(c.Request.Context(), err.Error())
//		c.JSON(http.StatusBadRequest, gin.H{
//			"success": false,
//			"message": "超时时间设置异常",
//		})
//		return
//	}
//
//	if chatModel.Stream {
//		c.Stream(func(w io.Writer) bool {
//			select {
//			case reply := <-replyChan:
//				timerReset(c, chatModel.Stream, timer, common.RequestOutTimeDuration)
//				urls := ""
//				if len(reply.EmbedUrls) > 0 {
//					for _, url := range reply.EmbedUrls {
//						urls += "\n" + fmt.Sprintf("![Image](%s)", url)
//					}
//				}
//				c.SSEvent("message", reply.Content+urls)
//				return true // 继续保持流式连接
//			case <-timer.C:
//				// 定时器到期时,关闭流
//				return false
//			case <-stopChan:
//				return false // 关闭流式连接
//			}
//		})
//	} else {
//		var replyResp model.ReplyResp
//		for {
//			select {
//			case reply := <-replyChan:
//				replyResp.Content = reply.Content
//				replyResp.EmbedUrls = reply.EmbedUrls
//			case <-timer.C:
//				c.JSON(http.StatusOK, gin.H{
//					"success": false,
//					"message": "request_out_time",
//				})
//				return
//			case <-stopChan:
//				c.JSON(http.StatusOK, gin.H{
//					"success": true,
//					"data":    replyResp,
//				})
//				return
//			}
//		}
//	}
//}

// ChatForOpenAI 发送消息-openai
// @Summary 发送消息-openai
// @Description 发送消息-openai
// @Tags openai
// @Accept json
// @Produce json
// @Param request body model.OpenAIChatCompletionRequest true "request"
// @Param Authorization header string false "Authorization"
// @Param out-time header string false "out-time"
// @Success 200 {object} model.OpenAIChatCompletionResponse "Successful response"
// @Router /v1/chat/completions [post]
func ChatForOpenAI(c *gin.Context) {

	var request model.OpenAIChatCompletionRequest
	err := json.NewDecoder(c.Request.Body).Decode(&request)
	if err != nil {
		common.LogError(c.Request.Context(), err.Error())
		c.JSON(http.StatusOK, model.OpenAIErrorResponse{
			OpenAIError: model.OpenAIError{
				Message: "无效的参数",
				Type:    "invalid_request_error",
				Code:    "invalid_parameter",
			},
		})
		return
	}

	sendChannelId, calledCozeBotId, err := getSendChannelIdAndCozeBotId(c, request.Model, true, request)
	common.SysLog(fmt.Sprintf("模型：{%s}，发送的机器人id:{%s}", request.Model, calledCozeBotId))
	content := "Hi！"
	messages := request.Messages

	for i := len(messages) - 1; i >= 0; i-- {
		message := messages[i]
		if message.Role == "user" {
			switch contentObj := message.Content.(type) {
			case string:
				//jsonData, err := json.Marshal(messages)
				//if err != nil {
				//	c.JSON(http.StatusOK, gin.H{
				//		"success": false,
				//		"message": err.Error(),
				//	})
				//	return
				//}
				//content = string(jsonData)
				content = contentObj
			case []interface{}:
				content, err = buildOpenAIGPT4VForImageContent(sendChannelId, contentObj)
				if err != nil {
					c.JSON(http.StatusOK, gin.H{
						"success": false,
						"message": err.Error(),
					})
					return
				}
			default:
				c.JSON(http.StatusOK, model.OpenAIErrorResponse{
					OpenAIError: model.OpenAIError{
						Message: "消息格式异常",
						Type:    "invalid_request_error",
						Code:    "discord_request_err",
					},
				})
				return

			}
			break
		}
	}

	if err != nil {
		common.LogError(c.Request.Context(), err.Error())
		c.JSON(http.StatusOK, model.OpenAIErrorResponse{
			OpenAIError: model.OpenAIError{
				Message: "配置异常",
				Type:    "invalid_request_error",
				Code:    "discord_request_err",
			},
		})
		return
	}

	sentMsg, err := discord.SendMessage(c, sendChannelId, calledCozeBotId, content)
	if err != nil {
		c.JSON(http.StatusOK, model.OpenAIErrorResponse{
			OpenAIError: model.OpenAIError{
				Message: err.Error(),
				Type:    "invalid_request_error",
				Code:    "discord_request_err",
			},
		})
		return
	}

	replyChan := make(chan model.OpenAIChatCompletionResponse)
	discord.RepliesOpenAIChans[sentMsg.ID] = replyChan
	defer delete(discord.RepliesOpenAIChans, sentMsg.ID)

	stopChan := make(chan string)
	discord.ReplyStopChans[sentMsg.ID] = stopChan
	defer delete(discord.ReplyStopChans, sentMsg.ID)

	timeDuration := common.RequestOutTimeDuration
	var isfastTime bool = common.Contains(common.FastModel, request.Model)
	var isPic bool = common.Contains(common.DrawMessages, content)
	if isPic {
		timeDuration = 1 * time.Minute
		common.SysLog(fmt.Sprintf("请求画图的，超时时间改为：{%s}", timeDuration))
	} else if isfastTime {
		timeDuration = 5 * time.Second
	}

	timer, err := setTimerWithHeader(c, request.Stream, timeDuration)
	if err != nil {
		common.LogError(c.Request.Context(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "超时时间设置异常",
		})
		return
	}

	if request.Stream {
		strLen := ""
		c.Stream(func(w io.Writer) bool {
			select {
			case reply := <-replyChan:
				timerReset(c, request.Stream, timer, timeDuration)
				//common.SysLog(fmt.Sprintf("响应信息：{%s}", reply.Choices[0].Message.Content))
				// TODO 多张图片问题
				if !strings.HasPrefix(reply.Choices[0].Message.Content, strLen) {
					if len(strLen) > 3 && strings.HasPrefix(reply.Choices[0].Message.Content, "\\n1.") {
						strLen = strLen[:len(strLen)-2]
					} else {
						return true
					}
				}

				newContent := strings.Replace(reply.Choices[0].Message.Content, strLen, "", 1)
				if newContent == "" && strings.HasSuffix(newContent, "[DONE]") {
					return true
				}
				common.SysLog(fmt.Sprintf("newContent响应信息：{%s}", newContent))
				reply.Choices[0].Delta.Content = newContent
				strLen += newContent

				reply.Object = "chat.completion.chunk"
				var isErrorMessage bool = common.SliceContains(common.CozeErrorMessages, reply.Choices[0].Message.Content)
				if isErrorMessage {
					reply.Object = "chat.error"
				}
				bytes, _ := common.Obj2Bytes(reply)
				c.SSEvent("", " "+string(bytes))

				if isErrorMessage {
					if common.SliceContains(common.CozeDailyLimitErrorMessages, reply.Choices[0].Message.Content) {
						common.LogWarn(c, fmt.Sprintf("USER_AUTHORIZATION: DAILY LIMIT"))
						//c.JSON(http.StatusOK, model.OpenAIErrorResponse{
						//	OpenAIError: model.OpenAIError{
						//		Message: reply.Choices[0].Message.Content,
						//		Type:    "model_response_error",
						//		Code:    "model_response_error",
						//	},
						//})
						//return false
					}
					common.LogWarn(c, "报错了："+reply.Choices[0].Message.Content)
					//discord.SetChannelDeleteTimer(sendChannelId, 5*time.Second)
					c.SSEvent("", " [DONE]")
					return false // 关闭流式连接
				}

				return true // 继续保持流式连接
			case <-timer.C:
				// 定时器到期时,关闭流
				contentTimeOut := "模型响应超时"
				if isPic {
					contentTimeOut = "图片模型响应超时"
				}
				c.JSON(http.StatusOK, model.OpenAIErrorResponse{
					OpenAIError: model.OpenAIError{
						Message: contentTimeOut,
						Type:    "model_response_timeout",
						Code:    "model_response_timeout",
					},
				})
				//c.SSEvent("", " [DONE]")
				return false
			case <-stopChan:
				c.SSEvent("", " [DONE]")
				return false // 关闭流式连接
			}
		})
	} else {
		var replyResp model.OpenAIChatCompletionResponse
		for {
			select {
			case reply := <-replyChan:
				replyResp = reply
			case <-timer.C:
				c.JSON(http.StatusOK, model.OpenAIErrorResponse{
					OpenAIError: model.OpenAIError{
						Message: "请求超时",
						Type:    "request_error",
						Code:    "request_out_time",
					},
				})
				return
			case <-stopChan:
				c.JSON(http.StatusOK, replyResp)
				return
			}
		}
	}
}

func buildOpenAIGPT4VForImageContent(sendChannelId string, objs []interface{}) (string, error) {
	var content string

	for i, obj := range objs {

		jsonData, err := json.Marshal(obj)
		if err != nil {
			return "", err
		}

		var req model.OpenAIGPT4VImagesReq
		err = json.Unmarshal(jsonData, &req)
		if err != nil {
			return "", err
		}

		if i == 0 && req.Type == "text" {
			content += req.Text
			continue
		} else if i == 1 && req.Type == "image_url" {
			if common.IsURL(req.ImageURL.URL) {
				content += fmt.Sprintf("\n%s", req.ImageURL.URL)
			} else if common.IsImageBase64(req.ImageURL.URL) {
				_, err := discord.UploadToDiscordAndGetURL(sendChannelId, req.ImageURL.URL)
				if err != nil {
					return "", fmt.Errorf("文件上传异常")
				}
				//content += fmt.Sprintf("\n%s", url)
			} else {
				return "", fmt.Errorf("文件格式有误")
			}
		} else {
			return "", fmt.Errorf("消息格式错误")
		}
	}
	//if runeCount := len([]rune(content)); runeCount > 2000 {
	//	return "", fmt.Errorf("prompt最大为2000字符 [%v]", runeCount)
	//}
	return content, nil

}

// ImagesForOpenAI 图片生成-openai
// @Summary 图片生成-openai
// @Description 图片生成-openai
// @Tags openai
// @Accept json
// @Produce json
// @Param request body model.OpenAIImagesGenerationRequest true "request"
// @Param Authorization header string false "Authorization"
// @Param out-time header string false "out-time"
// @Success 200 {object} model.OpenAIImagesGenerationResponse "Successful response"
// @Router /v1/images/generations [post]
func ImagesForOpenAI(c *gin.Context) {

	var request model.OpenAIImagesGenerationRequest
	err := json.NewDecoder(c.Request.Body).Decode(&request)
	if err != nil {
		common.LogError(c.Request.Context(), err.Error())
		c.JSON(http.StatusOK, model.OpenAIErrorResponse{
			OpenAIError: model.OpenAIError{
				Message: "无效的参数",
				Type:    "invalid_request_error",
				Code:    "invalid_parameter",
			},
		})
		return
	}

	if runeCount := len([]rune(request.Prompt)); runeCount > 2000 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("prompt最大为2000字符 [%v]", runeCount),
		})
		return
	}

	sendChannelId, calledCozeBotId, err := getSendChannelIdAndCozeBotId(c, request.Model, true, request)
	if err != nil {
		common.LogError(c.Request.Context(), err.Error())
		c.JSON(http.StatusOK, model.OpenAIErrorResponse{
			OpenAIError: model.OpenAIError{
				Message: "配置异常",
				Type:    "invalid_request_error",
				Code:    "discord_request_err",
			},
		})
		return
	}

	sentMsg, err := discord.SendMessage(c, sendChannelId, calledCozeBotId, request.Prompt)
	if err != nil {
		c.JSON(http.StatusOK, model.OpenAIErrorResponse{
			OpenAIError: model.OpenAIError{
				Message: err.Error(),
				Type:    "invalid_request_error",
				Code:    "discord_request_err",
			},
		})
		return
	}

	replyChan := make(chan model.OpenAIImagesGenerationResponse)
	discord.RepliesOpenAIImageChans[sentMsg.ID] = replyChan
	defer delete(discord.RepliesOpenAIImageChans, sentMsg.ID)

	stopChan := make(chan string)
	discord.ReplyStopChans[sentMsg.ID] = stopChan
	defer delete(discord.ReplyStopChans, sentMsg.ID)

	timer, err := setTimerWithHeader(c, false, common.RequestOutTimeDuration)
	if err != nil {
		common.LogError(c.Request.Context(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "超时时间设置异常",
		})
		return
	}

	var replyResp model.OpenAIImagesGenerationResponse
	for {
		select {
		case reply := <-replyChan:
			replyResp = reply
		case <-timer.C:
			c.JSON(http.StatusOK, model.OpenAIErrorResponse{
				OpenAIError: model.OpenAIError{
					Message: "请求超时",
					Type:    "request_error",
					Code:    "request_out_time",
				},
			})
			return
		case <-stopChan:
			if replyResp.Data == nil {
				c.JSON(http.StatusOK, model.OpenAIErrorResponse{
					OpenAIError: model.OpenAIError{
						Message: "discord未返回URL,检查prompt中是否有敏感内容",
						Type:    "invalid_request_error",
						Code:    "discord_request_err",
					},
				})
				return
			}
			c.JSON(http.StatusOK, replyResp)
			return
		}
	}

}

func getSendChannelIdAndCozeBotId(c *gin.Context, model string, isOpenAIAPI bool, request model.ChannelIdentifier) (sendChannelId string, calledCozeBotId string, err error) {
	secret := ""
	if isOpenAIAPI {
		if secret = c.Request.Header.Get("Authorization"); secret != "" {
			secret = strings.Replace(secret, "Bearer ", "", 1)
		}
	} else {
		secret = c.Request.Header.Get("proxy-secret")
	}

	//if secret == "" {
	//	if request.GetChannelId() == nil || *request.GetChannelId() == "" {
	//		return discord.ChannelId, discord.CozeBotId, nil
	//	} else {
	//		return *request.GetChannelId(), discord.CozeBotId, nil
	//	}
	//}

	channelCreateId := request.GetChannelId()
	if channelCreateId == nil || *channelCreateId == "" {
		channelCreateId = &discord.ChannelId
	}

	// botConfigs不为空
	if len(discord.BotConfigList) != 0 {

		botConfigs := discord.FilterConfigs(discord.BotConfigList, secret, model, nil)
		if len(botConfigs) != 0 {
			// 有值则随机一个
			botConfig, err := common.RandomElement(botConfigs)
			if err != nil {
				return "", "", err
			}
			//var sendChannelId string
			//sendChannelId, _ = discord.ChannelCreate(discord.GuildId, fmt.Sprintf("对话%s", c.Request.Context().Value(common.RequestIdKey)), 0)
			//discord.SetChannelDeleteTimer(sendChannelId, 5*time.Minute)
			return *channelCreateId, botConfig.CozeBotId, nil
		}
		// 使用原来的
		return *channelCreateId, discord.CozeBotId, nil
	} else {
		//channelCreateId, _ := discord.ChannelCreate(discord.GuildId, fmt.Sprintf("对话%s", c.Request.Context().Value(common.RequestIdKey)), 0)
		//discord.SetChannelDeleteTimer(*channelCreateId, 5*time.Minute)
		return *channelCreateId, discord.CozeBotId, nil
	}
}

func setTimerWithHeader(c *gin.Context, isStream bool, defaultTimeout time.Duration) (*time.Timer, error) {

	outTimeStr := getOutTimeStr(c, isStream)

	if outTimeStr != "" {
		outTime, err := strconv.ParseInt(outTimeStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return time.NewTimer(time.Duration(outTime) * time.Second), nil
	}
	return time.NewTimer(defaultTimeout), nil
}

func getOutTimeStr(c *gin.Context, isStream bool) string {
	var outTimeStr string
	if outTime := c.GetHeader(common.OutTime); outTime != "" {
		outTimeStr = outTime
	} else {
		if isStream {
			outTimeStr = common.StreamRequestOutTime
		} else {
			outTimeStr = common.RequestOutTime
		}
	}
	return outTimeStr
}

func timerReset(c *gin.Context, isStream bool, timer *time.Timer, defaultTimeout time.Duration) error {

	outTimeStr := getOutTimeStr(c, isStream)

	if outTimeStr != "" {
		outTime, err := strconv.ParseInt(outTimeStr, 10, 64)
		if err != nil {
			return err
		}
		timer.Reset(time.Duration(outTime) * time.Second)
		return nil
	}
	timer.Reset(defaultTimeout)
	return nil
}
