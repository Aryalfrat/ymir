package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IndustryEssentials/ymir-viewer/common/constants"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateViewer(t *testing.T) {
	server := NewViewerServer(constants.Config{ViewerURI: "127.0.0.1:9527"})

	server.getInt("invalid")
	server.getIntSliceFromString("")
	server.getIntSliceFromString(",")
	server.getIntSliceFromString("getIntSliceFromQuery")

	go server.Start()
	server.Clear()
}

type MockViewerHandler struct {
	mock.Mock
}

func (h *MockViewerHandler) GetAssetsHandler(
	mirRepo *constants.MirRepo,
	offset int,
	limit int,
	classIDs []int,
	currentAssetID string,
	cmTypes []int,
	cks []string,
	tags []string,
) constants.QueryAssetsResult {
	args := h.Called(mirRepo, offset, limit, classIDs, currentAssetID, cmTypes, cks, tags)
	return args.Get(0).(constants.QueryAssetsResult)
}

func (h *MockViewerHandler) GetDatasetDupHandler(
	mirRepo0 *constants.MirRepo,
	mirRepo1 *constants.MirRepo,
) constants.QueryDatasetDupResult {
	args := h.Called(mirRepo0, mirRepo1)
	return args.Get(0).(constants.QueryDatasetDupResult)
}

func (h *MockViewerHandler) GetDatasetMetaCountsHandler(
	mirRepo *constants.MirRepo,
) constants.QueryDatasetStatsResult {
	args := h.Called(mirRepo)
	return args.Get(0).(constants.QueryDatasetStatsResult)
}

func (h *MockViewerHandler) GetDatasetStatsHandler(
	mirRepo *constants.MirRepo,
	classIDs []int,
) constants.QueryDatasetStatsResult {
	args := h.Called(mirRepo, classIDs)
	return args.Get(0).(constants.QueryDatasetStatsResult)
}

func buildResponseBody(
	code constants.ResponseCode,
	msg constants.ResponseMsg,
	success bool,
	result interface{},
) []byte {
	resp := &ResultVO{Code: code, Msg: msg, Success: success, Result: result}
	bytes, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	return bytes
}

func TestStatsPageHandlerSuccess(t *testing.T) {
	mockHandler := MockViewerHandler{}
	viewer := &ViewerServer{gin: gin.Default(), handler: &mockHandler}

	r := viewer.gin
	r.GET("/users/:userID/repo/:repoID/branch/:branchID/dataset_stats", viewer.handleDatasetStats)

	userID := "userID"
	repoID := "repoID"
	branchID := "branchID"
	classIDs := []int{0, 1}
	classIDsStr := "0,1"
	statsRequestURL := fmt.Sprintf(
		"/users/%s/repo/%s/branch/%s/dataset_stats?class_ids=%s",
		userID,
		repoID,
		branchID,
		classIDsStr,
	)

	statsExpectedResult := constants.NewQueryDatasetStatsResult()
	for classID := range classIDs {
		statsExpectedResult.Gt.ClassIdsCount[classID] = 0
		statsExpectedResult.Pred.ClassIdsCount[classID] = 0
	}
	statsExpectedResponseData := buildResponseBody(
		constants.ViewerSuccessCode,
		"Success",
		true,
		statsExpectedResult,
	)

	mirRepo := constants.MirRepo{UserID: userID, RepoID: repoID, BranchID: branchID, TaskID: branchID}
	mockHandler.On("GetDatasetStatsHandler", &mirRepo, classIDs).Return(statsExpectedResult)

	req, _ := http.NewRequest("GET", statsRequestURL, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, string(statsExpectedResponseData), w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMetaCountPageHandlerSuccess(t *testing.T) {
	mockHandler := MockViewerHandler{}
	viewer := &ViewerServer{gin: gin.Default(), handler: &mockHandler}

	r := viewer.gin
	r.GET("/users/:userID/repo/:repoID/branch/:branchID/dataset_meta_count", viewer.handleDatasetMetaCounts)

	userID := "userID"
	repoID := "repoID"
	branchID := "branchID"
	metaRequestURL := fmt.Sprintf(
		"/users/%s/repo/%s/branch/%s/dataset_meta_count",
		userID,
		repoID,
		branchID,
	)

	metaExpectedResult := constants.NewQueryDatasetStatsResult()
	metaExpectedResponseData := buildResponseBody(
		constants.ViewerSuccessCode,
		"Success",
		true,
		metaExpectedResult,
	)

	mirRepo := constants.MirRepo{UserID: userID, RepoID: repoID, BranchID: branchID, TaskID: branchID}
	mockHandler.On("GetDatasetMetaCountsHandler", &mirRepo).Return(metaExpectedResult)

	req, _ := http.NewRequest("GET", metaRequestURL, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, string(metaExpectedResponseData), w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMetaCountPageHandlerFailure(t *testing.T) {
	mockHandler := MockViewerHandler{}
	viewer := &ViewerServer{gin: gin.Default(), handler: &mockHandler}

	r := viewer.gin
	r.GET("/users/:userID/repo/:repoID/branch/:branchID/dataset_meta_count", viewer.handleDatasetMetaCounts)

	userID := "userID"
	repoID := "repoID"
	branchID := "branchID"
	metaRequestURL := fmt.Sprintf(
		"/users/%s/repo/%s/branch/%s/dataset_meta_count",
		userID,
		repoID,
		branchID,
	)

	failureResult := FailureResult{
		Code: constants.FailRepoNotExistCode,
		Msg:  constants.ResponseMsg("unknown ref"),
	}
	statsExpectedResponseData := buildResponseBody(
		failureResult.Code,
		failureResult.Msg,
		false,
		failureResult,
	)

	mirRepo := constants.MirRepo{UserID: userID, RepoID: repoID, BranchID: branchID, TaskID: branchID}
	mockHandler.On("GetDatasetMetaCountsHandler", &mirRepo).Panic("unknown ref")

	req, _ := http.NewRequest("GET", metaRequestURL, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, string(statsExpectedResponseData), w.Body.String())
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDupPageHandlerSuccess(t *testing.T) {
	mockHandler := MockViewerHandler{}
	viewer := &ViewerServer{gin: gin.Default(), handler: &mockHandler}

	r := viewer.gin
	r.GET("/users/:userID/repo/:repoID/dataset_duplication", viewer.handleDatasetDup)

	userID := "userID"
	repoID := "repoID"
	branchID0 := "branchID0"
	branchID1 := "branchID1"
	dupRequestURL := fmt.Sprintf(
		"/users/%s/repo/%s/dataset_duplication?candidate_dataset_ids=%s,%s",
		userID,
		repoID,
		branchID0,
		branchID1,
	)

	dupCount := 100
	branchCount0 := int64(1000)
	branchCount1 := int64(2000)
	mockDupResult := constants.QueryDatasetDupResult{
		Duplication: dupCount,
		TotalCount:  map[string]int64{branchID0: branchCount0, branchID1: branchCount1},
	}
	statsExpectedResponseData := buildResponseBody(
		constants.ViewerSuccessCode,
		"Success",
		true,
		mockDupResult,
	)

	// Set mock funcs.
	mirRepo0 := constants.MirRepo{UserID: userID, RepoID: repoID, BranchID: branchID0, TaskID: branchID0}
	mirRepo1 := constants.MirRepo{UserID: userID, RepoID: repoID, BranchID: branchID1, TaskID: branchID1}
	mockHandler.On("GetDatasetDupHandler", &mirRepo0, &mirRepo1).Return(mockDupResult)

	req, _ := http.NewRequest("GET", dupRequestURL, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, string(statsExpectedResponseData), w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDupPageHandlerFailure(t *testing.T) {
	mockHandler := MockViewerHandler{}
	viewer := &ViewerServer{gin: gin.Default(), handler: &mockHandler}

	r := viewer.gin
	r.GET("/users/:userID/repo/:repoID/dataset_duplication", viewer.handleDatasetDup)

	userID := "userID"
	repoID := "repoID"
	dupRequestURL0 := fmt.Sprintf(
		"/users/%s/repo/%s/dataset_duplication",
		userID,
		repoID,
	)
	failureResult := FailureResult{
		Code: constants.FailInvalidParmsCode,
		Msg:  constants.ResponseMsg("Invalid candidate_dataset_ids."),
	}
	statsExpectedResponseData := buildResponseBody(
		failureResult.Code,
		failureResult.Msg,
		false,
		failureResult,
	)
	req, _ := http.NewRequest("GET", dupRequestURL0, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, string(statsExpectedResponseData), w.Body.String())
	assert.Equal(t, http.StatusBadRequest, w.Code)

	branchID0 := "branchID0"
	dupRequestURL1 := fmt.Sprintf(
		"/users/%s/repo/%s/dataset_duplication?candidate_dataset_ids=%s",
		userID,
		repoID,
		branchID0,
	)
	failureResult = FailureResult{
		Code: constants.FailInvalidParmsCode,
		Msg:  constants.ResponseMsg("candidate_dataset_ids requires exact two datasets."),
	}
	statsExpectedResponseData = buildResponseBody(
		failureResult.Code,
		failureResult.Msg,
		false,
		failureResult,
	)
	req, _ = http.NewRequest("GET", dupRequestURL1, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, string(statsExpectedResponseData), w.Body.String())
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssetsPageHandlerSuccess(t *testing.T) {
	mockHandler := MockViewerHandler{}
	viewer := &ViewerServer{gin: gin.Default(), handler: &mockHandler}

	r := viewer.gin
	r.GET("/users/:userID/repo/:repoID/branch/:branchID/assets", viewer.handleAssets)

	userID := "userID"
	repoID := "repoID"
	branchID := "branchID"
	offset := -1
	limit := 0
	classIDs := []int{0, 1}
	classIDsStr := "0,1"
	currentAssetID := "asset_id"
	cmTypes := "0,1"
	cks := "ck0,ck1"
	tags := "tag0,tag1"
	querySuffix := fmt.Sprintf("offset=%d&limit=%d&class_ids=%s&current_asset_id=%s&cm_types=%s&cks=%s&tags=%s",
		offset,
		limit,
		classIDsStr,
		currentAssetID,
		cmTypes,
		cks,
		tags,
	)
	dupRequestURL := fmt.Sprintf(
		"/users/%s/repo/%s/branch/%s/assets?%s",
		userID,
		repoID,
		branchID,
		querySuffix,
	)

	assetsExpectedResult := constants.QueryAssetsResult{
		AssetsDetail:     []constants.MirAssetDetail{},
		Offset:           0,
		Limit:            1,
		Anchor:           int64(len(classIDs)),
		TotalAssetsCount: 42,
	}
	assetsExpectedResponseData := buildResponseBody(
		constants.ViewerSuccessCode,
		"Success",
		true,
		assetsExpectedResult,
	)

	revisedOffset := 0
	revisedLimit := 1
	revisedcmTypes := []int{0, 1}
	revisedCks := []string{"ck0", "ck1"}
	revisedTags := []string{"tag0", "tag1"}
	mirRepo := constants.MirRepo{UserID: userID, RepoID: repoID, BranchID: branchID, TaskID: branchID}
	mockHandler.On(
		"GetAssetsHandler",
		&mirRepo,
		revisedOffset,
		revisedLimit,
		classIDs,
		currentAssetID,
		revisedcmTypes,
		revisedCks,
		revisedTags).
		Return(assetsExpectedResult)

	req, _ := http.NewRequest("GET", dupRequestURL, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, string(assetsExpectedResponseData), w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
}
