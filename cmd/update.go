package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TCP404/OneTiny-cli/common/define"
	"github.com/TCP404/OneTiny-cli/config"
	"github.com/TCP404/OneTiny-cli/internal/model"

	"github.com/TCP404/eutil"
	"github.com/fatih/color"
	"github.com/parnurzeal/gorequest"

	"github.com/urfave/cli/v2"
)

var updateCmd = newUpdateCmd()

func newUpdateCmd() *cli.Command {
	return &cli.Command{
		Name:    "update",
		Aliases: []string{"u", "up"},
		Usage:   "更新 OneTiny 到最新版",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "list",
				Aliases:     []string{"l"},
				Usage:       "列出远程服务器上所有可用版本",
				Required:    false,
				DefaultText: "false",
			},
			&cli.StringFlag{
				Name:     "use",
				Usage:    "指定版本号",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			var u = &update{currVersion: splitVersion(define.VERSION)}
			err := u.updateAction(c)
			if err != nil {
				return cli.Exit(err.Error(), 31)
			}
			return cli.Exit(u.msg, 0)
		},
	}
}

type update struct {
	currVersion []string
	msg         string
}

func (u *update) updateAction(c *cli.Context) error {
	switch {
	case c.IsSet("list"):
		return u.updateList()
	case c.IsSet("use"):
		return u.updateVersion(c.String("use"))
	}
	return u.updateLatest()
}

func (u *update) updateList() error {
	tags, err := getVersionList()
	if err != nil {
		return err
	}
	for _, tag := range tags {
		fmt.Println(color.GreenString("%v", tag.TagName))
	}
	return nil
}

func (u *update) updateVersion(version string) error {
	if err := checkVersion(version); err != nil {
		return err
	}

	req := gorequest.New()
	_, body, errs := req.Get(define.VersionByTagURL + version).End()
	if len(errs) != 0 {
		return errors.New("网络抖动了一下～请重试")
	}
	var versionInfo = new(model.ReleaseInfo)
	err := json.Unmarshal([]byte(body), versionInfo)
	if err != nil {
		return errors.New("网络抖动了一下～请重试")
	}

	// 检查当前系统是 linux 还是 mac 还是 windows决定 Assets 用哪个,然后进行下载
	name, ok := define.ReleaseName[config.OS]
	if !ok {
		u.msg = color.YellowString("暂时没有适合您的系统的版本，请自行下载编译")
		return nil
	}
	var (
		assert *model.ReleaseAsset
		l      = len(versionInfo.Assets)
	)
	for i := 0; i < l; i++ {
		if versionInfo.Assets[i].Name == name {
			assert = &versionInfo.Assets[i]
			break
		}
	}
	if assert == nil {
		u.msg = color.YellowString("暂时没有适合您的系统的版本，请自行下载编译")
		return nil
	}

	// 进行下载
	p, err := os.UserHomeDir()
	if err != nil {
		u.msg = color.HiYellowString("获取 Home 目录失败")
		p = config.Pwd
	}
	path := filepath.Join(p, assert.Name)
	errs = eutil.DownloadBinary(assert.DownloadURL, path)
	if len(errs) != 0 {
		return errors.New("网络抖动了一下～请重试")
	}
	u.msg += color.HiGreenString("更新完成～, 文件存放于: %s", path)
	return nil
}

func (u *update) updateLatest() error {
	// 获取当前最新版本
	req := gorequest.New()
	_, body, errs := req.Get(define.VersionLatestURL).End()
	if len(errs) != 0 {
		return errors.New("网络抖动了一下～请重试")
	}

	var latestInfo = new(model.ReleaseInfo)
	err := json.Unmarshal([]byte(body), latestInfo)
	if err != nil {
		return errors.New("网络抖动了一下～请重试")
	}

	// 检查最新版本与当前版本
	latestVersion := splitVersion(latestInfo.TagName)
	if u.isLatest(latestVersion) {
		u.msg = color.GreenString("当前已是最新版本~")
		return nil
	}
	// 进行更新
	return u.updateVersion(latestInfo.TagName)
}

func (u *update) isLatest(version []string) bool {
	max := len(version)
	if max > len(u.currVersion) {
		max = len(u.currVersion)
	}

	for i := 0; i < max; i++ {
		if version[i] < u.currVersion[i] {
			return false
		}
	}
	return true
}

func splitVersion(version string) (v []string) {
	var major, minor, revision = "0", "0", "0"
	sArr := strings.Split(strings.TrimLeft(version, "v"), ".")
	if len(sArr) >= 3 {
		revision = sArr[2]
	}
	if len(sArr) >= 2 {
		minor = sArr[1]
	}
	if len(sArr) >= 1 {
		major = sArr[0]
	}
	return append(v, major, minor, revision)
}

func getVersionList() ([]model.TagList, error) {
	_, body, errs := gorequest.New().Set("Accept", "application/vnd.github.v3+json").Get(define.VersionListURL).End()
	if len(errs) != 0 {
		return nil, errors.New("网络抖动了一下～请重试")
	}
	var tags []model.TagList
	err := json.Unmarshal([]byte(body), &tags)
	if err != nil {
		return nil, errors.New("网络抖动了一下～请重试")
	}
	return tags, nil
}

func checkVersion(version string) error {
	tags, err := getVersionList()
	if err != nil {
		return err
	}
	for _, tag := range tags {
		if strings.Compare(tag.TagName, version) == 0 {
			return nil
		}
	}
	return errors.New("找不到您指定的版本，请检查您输入的版本号")
}
