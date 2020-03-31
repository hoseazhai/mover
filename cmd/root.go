/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"fmt"
	"github.com/hoseazhai/mover/cloud"
	"github.com/hoseazhai/mover/commons/utils"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	postPath string
)

//const signImg = "sinaimg.cn"
const signImg = "png"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mover",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mover.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().StringVarP(&postPath, "post", "p", "./", "指定markdown文件路径")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".mover" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".mover")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// StartMove ... 开始执行迁移程序
func StartMover(uploader cloud.Uploader) {
	// 获取所以的md文件
	files, err := utils.GetAllFiles(postPath)
	if err != nil {
		fmt.Printf("获取markdown文件出错了：%v\n", err)
		return
	}
	for _, file := range files {
		err := parseFile(file, uploader)
		if err != nil {
			fmt.Printf("解析文件：%s 出错了：%v \n", file, err)
			continue
		}
	}
}

func parseFile(filePath string, uploader cloud.Uploader) error {
	bt, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	sep := string(os.PathSeparator)
	// 获取文件内容
	content := string(bt)
	reg := regexp.MustCompile(`!\[.*?\]\((.*?)\)|<img.*?src=[\'\"](.*?)[\'\"].*?>`)
	params := reg.FindAllStringSubmatch(content, -1)
	// 获取所以的图片
	for _, param := range params {
		imgURL := param[1]
		// 微博图床才处理
		if strings.Index(imgURL, signImg) != -1 && strings.Index(imgURL, "http") == -1 {
			absUrl := strings.Replace(filePath, "\\", "/", -1)
			fmt.Println(filePath)

			imgName := path.Base(imgURL)
			relPath := path.Dir(imgURL)
			absPath := path.Dir(absUrl)

			tempPath := absPath + sep + relPath
			imgPath, _ := filepath.Abs(tempPath)

			imgParamsURL := fmt.Sprintf("images/%s", imgName)
			url := imgPath + sep + imgName
			f, err := ioutil.ReadFile(url)
			if err != nil {
				fmt.Println("Error: 读取文件", err)
				os.Exit(-1)
			}
			cloudURL := fmt.Sprintf("https://%s.%s/%s", ossBucket, ossEndpoint, imgParamsURL)
			key, err := uploader.Uploader(imgParamsURL, bytes.NewReader(f))
			fmt.Println("objectKey", key, "==========")

			//cloudURL, err := uploadToCloud(imgURL, uploader)
			if err != nil {
				fmt.Printf("图片： %s 转换出错：%v \n", cloudURL, err)
				continue
			}
			newContent := strings.Replace(content, imgURL, cloudURL, -1)
			// 重新写入文件
			ioutil.WriteFile(filePath, []byte(newContent), 0)
			content = newContent
			fmt.Printf("图片：%s 替换成功\n", imgURL)
		} else if strings.Index(imgURL, "http") != -1 {
			cloudURL, err := uploadToCloud(imgURL, uploader)
			if err != nil {
				fmt.Printf("图片： %s 转换出错：%v \n", cloudURL, err)
				continue
			}
			newContent := strings.Replace(content, imgURL, cloudURL, -1)
			// 重新写入文件
			ioutil.WriteFile(filePath, []byte(newContent), 0)
			content = newContent
			fmt.Printf("图片：%s 替换成功\n", imgURL)
		}
	}
	return nil
}

func uploadToCloud(url string, uploader cloud.Uploader) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	objectKey := fmt.Sprintf("images/%s", utils.RandID(8))
	key, err := uploader.Uploader(objectKey, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://%s.%s.%s", ossBucket, ossEndpoint, key), nil
}
