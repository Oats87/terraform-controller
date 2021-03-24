package routes

import (
	"bytes"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/jsonapi"
	"github.com/hashicorp/go-tfe"
	"github.com/rancher/terraform-controller/pkg/types"

	//"github.com/sirupsen/logrus"
	"compress/gzip"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var cs *types.Controllers

func Register(r *gin.Engine, controllers *types.Controllers) error {
	cs = controllers
	r.GET("/api/v2/ping", ping)
	r.GET("/.well-known/terraform.json", discovery)
	r.GET("/api/v2/organizations/:org/entitlement-set", entitlement)
	r.GET("/api/v2/organizations/:org/workspaces/:workspace", workspace)
	r.GET("/api/v2/workspaces/:workspace/current-state-version", state)
	r.GET("/api/v2/download/:workspace/state", stateDownload)
	r.POST("/api/v2/workspaces/:workspace/actions/lock", stateLock)
	r.POST("/api/v2/workspaces/:workspace/actions/unlock", stateUnlock)

	return nil
}

func ping(c *gin.Context) {
	c.String(200, "pong")
}

func entitlement(c *gin.Context) {
	c.Header("Content-Type", jsonapi.MediaType)
	ent := &tfe.Entitlements{
		Operations: true,
	}
	jsonapi.MarshalPayload(c.Writer, ent)
}

func discovery(c *gin.Context) {
	c.JSON(200, gin.H{
		"tfe.v2":   "/api/v2/",
		"tfe.v2.1": "/api/v2/",
		"tfe.v2.2": "/api/v2/",
	})
}

func workspace(c *gin.Context) {
	c.Header("Content-Type", jsonapi.MediaType)
	workspace := &tfe.Workspace{}
	workspace.Name = c.Param("workspace")
	workspace.ID = "ws-123"
	jsonapi.MarshalPayload(c.Writer, workspace)
}

func stateLock(c *gin.Context) {
	c.String(200, "OK")
}
func stateUnlock(c *gin.Context) {
	c.String(200, "OK")
}
func state(c *gin.Context) {
	c.Header("Content-Type", jsonapi.MediaType)
	ws := c.Param("workspace")

	stateVersion := &tfe.StateVersion{}
	stateVersion.DownloadURL = fmt.Sprintf("download/%s/state", ws)
	jsonapi.MarshalPayload(c.Writer, stateVersion)
}

func stateDownload(c *gin.Context) {
	c.Header("Content-Type", jsonapi.MediaType)
	//	ws := c.Param("workspace")
	secret, _ := cs.Secret.Get("default", "tfstate-default-my-state", metav1.GetOptions{})
	state, _ := gunzip(secret.Data["tfstate"])
	c.String(200, state)
}

func gunzip(data []byte) (string, error) {
	b := bytes.NewBuffer(data)
	var r io.Reader
	r, err := gzip.NewReader(b)
	if err != nil {
		return "", err
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return "", err
	}

	return string(resB.Bytes()), nil
}
