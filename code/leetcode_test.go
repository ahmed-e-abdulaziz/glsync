package code

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ahmed-e-abdulaziz/gh-leet-sync/config"
	"github.com/stretchr/testify/assert"
)

var recorder *httptest.ResponseRecorder
var testUrl string
var submissionListCalled = false
var submissionDetailsCalled = false
var userProgressQuestionListCalled = false

func TestMain(m *testing.M) {
	recorder = httptest.NewRecorder()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBodyBytes, _ := io.ReadAll(r.Body)
		reqBody := string(reqBodyBytes)
		if strings.Contains(reqBody, "submissionList") {
			submissionListCalled = true
			w.Write([]byte(
				`{
					"data": {
						"questionSubmissionList": {
							"lastKey": null,
							"hasNext": true,
							"submissions": [
								{
									"id": "1490835403",
									"title": "Longest Consecutive Sequence",
									"titleSlug": "longest-consecutive-sequence",
									"status": 10,
									"statusDisplay": "Accepted",
									"lang": "golang",
									"langName": "Go",
									"runtime": "55 ms",
									"timestamp": "1735406731",
									"url": "/submissions/detail/1490835403/",
									"isPending": "Not Pending",
									"memory": "11.7 MB",
									"hasNotes": false,
									"notes": "",
									"flagType": "WHITE",
									"frontendId": 2,
									"topicTags": []
								}
							]
						}
					}
				}`))
		}
		if strings.Contains(reqBody, "submissionDetails") {
			submissionDetailsCalled = true
			w.Write([]byte(
				`{
					"data": {
						"submissionDetails": {
							"runtime": 55,
							"runtimeDisplay": "55 ms",
							"runtimePercentile": 49.73569999999996,
							"runtimeDistribution": "{\"lang\": \"golang\", \"distribution\": [[\"14\", 9.696900000000003], [\"38\", 37.1373], [\"61\", 14.314100000000002], [\"85\", 3.034299999999999], [\"109\", 1.8472], [\"132\", 0.9897], [\"156\", 0.264], [\"179\", 0.39580000000000004], [\"203\", 0.33], [\"226\", 0.3299], [\"250\", 0.066], [\"274\", 0.066], [\"297\", 0.066], [\"321\", 0.198], [\"344\", 0.132], [\"368\", 0.066], [\"391\", 0.066], [\"415\", 0.066], [\"439\", 0.198], [\"462\", 0.32989999999999997], [\"486\", 0.3299], [\"509\", 0.132], [\"533\", 0.066], [\"557\", 0.066], [\"580\", 0.066], [\"604\", 0.066], [\"627\", 0.066], [\"651\", 0.066], [\"674\", 0.066], [\"698\", 0.066], [\"722\", 0.066], [\"745\", 0.066], [\"769\", 0.066], [\"792\", 0.1979], [\"816\", 0.1319], [\"839\", 0.1319], [\"863\", 0.066], [\"887\", 0.066], [\"910\", 0.066], [\"934\", 0.066], [\"957\", 0.066], [\"981\", 0.066], [\"1004\", 0.1979], [\"1028\", 0.066], [\"1052\", 0.1319], [\"1075\", 0.066], [\"1099\", 0.066], [\"1122\", 0.2639], [\"1146\", 0.1319], [\"1169\", 0.1319], [\"1193\", 0.066], [\"1217\", 0.1319], [\"1240\", 0.2639], [\"1264\", 0.1319], [\"1287\", 0.066], [\"1311\", 0.1319], [\"1334\", 0.1319], [\"1358\", 0.1979], [\"1382\", 0.1319], [\"1405\", 0.066], [\"1429\", 0.1979], [\"1452\", 0.1319], [\"1476\", 0.1319], [\"1500\", 0.066], [\"1523\", 0.1979], [\"1547\", 0.066], [\"1570\", 0.066], [\"1594\", 0.1319], [\"1617\", 0.066], [\"1641\", 0.066], [\"1665\", 0.066], [\"1688\", 0.1319], [\"1712\", 0.2639], [\"1735\", 0.3298], [\"1759\", 0.1979], [\"1782\", 1.8471], [\"1806\", 4.8153], [\"1830\", 4.3538], [\"1853\", 3.8919000000000006], [\"1877\", 3.9579999999999997]]}",
							"memory": 11704000,
							"memoryDisplay": "11.7 MB",
							"memoryPercentile": 53.42979999999998,
							"memoryDistribution": "{\"lang\": \"golang\", \"distribution\": [[\"9356\", 0.132], [\"9468\", 0.1319], [\"9581\", 0.1979], [\"9693\", 0.1319], [\"9806\", 0.1979], [\"9918\", 0.1979], [\"10031\", 0.1979], [\"10143\", 1.1873], [\"10256\", 2.8364000000000003], [\"10368\", 2.3747], [\"10481\", 1.847], [\"10593\", 1.847], [\"10706\", 1.781], [\"10818\", 2.9024], [\"10931\", 3.2982], [\"11043\", 3.8259], [\"11156\", 7.5198], [\"11268\", 4.4855], [\"11381\", 3.562], [\"11493\", 3.4301], [\"11606\", 4.4855], [\"11718\", 4.0237], [\"11831\", 4.2876], [\"11943\", 3.628], [\"12056\", 4.8812999999999995], [\"12168\", 1.6491], [\"12281\", 1.1873], [\"12393\", 1.4512], [\"12506\", 1.3193], [\"12618\", 0.6596], [\"12731\", 1.3193], [\"12843\", 0.9235], [\"12956\", 1.0554000000000001], [\"13068\", 0.9235], [\"13181\", 0.1979], [\"13293\", 0.5937], [\"13406\", 0.3298], [\"13518\", 0.5277], [\"13631\", 1.4512], [\"13743\", 1.9129], [\"13856\", 3.3640999999999996], [\"13968\", 0.8575], [\"14081\", 1.2533], [\"14193\", 0.7916], [\"14306\", 0.5937], [\"14418\", 0.4617], [\"14531\", 0.5277], [\"14643\", 0.5277], [\"14756\", 0.6596], [\"14868\", 0.4617], [\"14981\", 0.2639], [\"15093\", 0.3298], [\"15206\", 0.1979], [\"15318\", 0.066], [\"15431\", 0.1319], [\"15543\", 0.2639], [\"15656\", 0.7916], [\"15768\", 0.2639], [\"15881\", 0.3958], [\"15993\", 0.5277], [\"16106\", 0.3298], [\"16218\", 0.1979], [\"16331\", 0.1979], [\"16443\", 0.1319], [\"16556\", 0.3958], [\"16668\", 0.066], [\"16781\", 0.1979], [\"16893\", 0.2639], [\"17006\", 0.066], [\"17118\", 0.066], [\"17231\", 0.1319], [\"17343\", 0.1319], [\"17456\", 0.2638], [\"17568\", 0.1319], [\"17681\", 0.066], [\"17793\", 0.1319], [\"17906\", 0.1979], [\"18018\", 0.1979], [\"18131\", 0.066], [\"18243\", 0.1319]]}",
							"code": "func longestConsecutive(nums []int) int {\n\tset := map[int]struct{}{}\n\tfor _, num := range nums {\n\t\tset[num] = struct{}{}\n\t}\n\tmaxCount := 0\n\tfor i := 0; i < len(nums); i++ {\n\t\tcount := 1\n\t\tif _, ok := set[nums[i]]; ok {\n\t\t\tprev := nums[i] - 1\n\t\t\t_, hasPrev := set[prev]\n\t\t\tfor hasPrev {\n\t\t\t\tdelete(set, prev)\n\t\t\t\tprev--\n\t\t\t\tcount++\n\t\t\t\t_, hasPrev = set[prev]\n\t\t\t}\n\t\t\tnext := nums[i] + 1\n\t\t\t_, hasNext := set[next]\n\t\t\tfor hasNext {\n\t\t\t\tdelete(set, next)\n\t\t\t\tnext++\n\t\t\t\tcount++\n\t\t\t\t_, hasNext = set[next]\n\t\t\t}\n\t\t}\n\t\tmaxCount = max(count, maxCount)\n\t}\n\treturn maxCount\n}\n\nfunc max(x, y int) int {\n\tif x > y {\n\t\treturn x\n\t} else {\n\t\treturn y\n\t}\n}\n",
							"timestamp": 1735406731,
							"statusCode": 10,
							"user": {
								"username": "ahmedehab95",
								"profile": {
									"realName": "ahmedehab95",
									"userAvatar": "https://assets.leetcode.com/users/default_avatar.jpg"
								}
							},
							"lang": {
								"name": "golang",
								"verboseName": "Go"
							},
							"question": {
								"questionId": "128",
								"titleSlug": "longest-consecutive-sequence",
								"hasFrontendPreview": false
							},
							"notes": "",
							"flagType": "WHITE",
							"topicTags": [],
							"runtimeError": null,
							"compileError": null,
							"lastTestcase": "",
							"codeOutput": "",
							"expectedOutput": "",
							"totalCorrect": 77,
							"totalTestcases": 77,
							"fullCodeOutput": null,
							"testDescriptions": null,
							"testBodies": null,
							"testInfo": null,
							"stdOutput": ""
						}
					}
				}`))
		}
		if strings.Contains(reqBody, "userProgressQuestionList") {
			userProgressQuestionListCalled = true
			w.Write([]byte(
				`{
					"data": {
						"userProgressQuestionList": {
							"questions": [
								{
									"frontendId": "128",
									"title": "Longest Consecutive Sequence",
									"titleSlug": "longest-consecutive-sequence",
									"lastSubmittedAt": "2024-12-28T17:25:31+00:00",
									"questionStatus": "SOLVED",
									"lastResult": "AC"
								}
							]
						}
					}
				}`))
		}
	}))
	testUrl = "http://" + server.Listener.Addr().String()
	m.Run()
}

func TestFetchSubmissions(t *testing.T) {
	leetcode := NewLeetCode(config.Config{LCookie: "COOKIE", RepoUrl: "REPO_URL"}, testUrl)
	submission := leetcode.FetchSubmissions()[0]
	assert.Equal(t, submission.Id, "128")
	assert.Equal(t, submission.Lang, "golang")
	assert.Equal(t, submission.Title, "Longest Consecutive Sequence")
	assert.True(t, userProgressQuestionListCalled)
	assert.True(t, submissionListCalled)
	assert.True(t, submissionDetailsCalled)
}
