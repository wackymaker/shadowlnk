# shadowlnk介绍 / Introduction to shadowlnk
当我在尝试在免杀上尝试进一步进行持久化操作的时候遇到了阻碍,我发现大部分中国杀软（以及一些其他检查敏感行为比较严重的杀软）很难进行常规的三大维权行为（注册表，自动任务，自启动）的更改，于是我尝试了一个比较弯弯绕绕的方式构筑了这个小工具。

希望这个小工具能给大家带来灵感。

When I tried to perform persistence operations in the context of evading antivirus detection, I encountered obstacles,
I found that most Chinese antivirus software (as well as some other AVs that strictly monitor sensitive behaviors) make it very difficult to modify the three common persistence vectors: registry, scheduled tasks, and startup entries.
So I tried a rather roundabout method and created this small tool.

hoping it can inspire others.
## shadowlnk逻辑 / shadowlnk Logic

这个小东西的逻辑非常简单

它用于劫持某个快捷方式（或者自动劫持某个目录下所有快捷方式）

将它真正的目标绑定我们投放在此系统上的载荷

做到诱导目标人物在启动某个自己程序的时候自动上线

并且这一切流程是无感的 

The logic of this little tool is very simple

It is used to hijack a shortcut (or automatically hijack all shortcuts in a specified directory)

binding its original target to a payload we deploy on the system

so that when the target launches their own program, they automatically run our payload

and this entire process is stealthy and imperceptible

正常运作它需要接受三个最基础的操作：  

To work properly, it requires three basic parameters:

- `-i` 指定单个要劫持的快捷方式文件路径  
  `-i` Specify the path of a single shortcut file to hijack

- `-w` 可写目录（用于存放生成的隐藏目录和辅助文件）  
  `-w` Writable directory (used to store the generated hidden folder and helper files)

- `-p` 需要执行的载荷路径  
  `-p` The path of the payload to execute

这里我们模拟一个使用情况：  
Here is a sample usage:

```zsh
.\shadowlnk.exe -i "c:\user\administrator\desktop\firefox.lnk" -w "c:\test" -p "c:\test\beacon.exe"
```
shadowslnk会自动查找这个lnk指向的exe，并记录它的名称,之后在指定可写目录下创建一个同名的目录并添加上系统隐藏属性，这让一般没有开启显示隐藏文件夹的目标难以找到我们的中转文件夹，随后会生成三个文件

就这个例子为firefox.bat,firefox.vbs,firefox.ico,其中vbs文件负责不显示黑框运行bat文件，bat文件则先启动我们指定的载荷，在一个很短几乎无感的时间内启动原lnk指向目标程序

最后程序会自动读取目标程序的ico图标，它会用这个图标生成一个新lnk指向vbs脚本，并替换桌面上的lnk为我们新创建的lnk

shadowlnk will automatically locate the exe file targeted by the shortcut (lnk) and record its name. Then, it creates a folder with the same name in the specified writable directory and sets the system hidden attribute on it, making it difficult for targets who have not enabled showing hidden files and folders to find our intermediate folder.

For this example, three files are generated: firefox.bat, firefox.vbs, and firefox.ico. The vbs file runs the bat file without showing a command window, and the bat file first launches our specified payload, then quickly (almost imperceptibly) starts the original program targeted by the shortcut.

Finally, the program automatically reads the ico icon of the original target program, uses this icon to generate a new shortcut (lnk) pointing to the vbs script, and replaces the original desktop shortcut with our newly created one.

### 效果 / Effect
```zsh
C:\Users\wacky\Desktop\shadowlnk>.\shadowlnk_64.exe -i "C:\Users\wacky\Desktop\firefox.exe - 快捷方式.lnk" -w "C:\Users\wacky\Desktop\shadowlnk" -p "C:\Users\wacky\Desktop\shadowlnk\beacon_x64.exe" -power
处理快捷方式 / Processing shortcut: C:\Users\wacky\Desktop\firefox.exe - 快捷方式.lnk
原目标 / Original target: C:\Program Files\Mozilla Firefox\firefox.exe
劫持成功 / Hijack successful: C:\Users\wacky\Desktop\firefox.exe - 快捷方式.lnk

C:\Users\wacky\Desktop\shadowlnk>

```
简单来说经过极短的转换，目标桌面的快捷方式就会被替换成我们的劫持版本，此后用户每次点击此lnk运行都会上线我们的载荷，并且几乎无感

In short, after a very brief transformation, the target’s desktop shortcut will be replaced by our hijacked version. Thereafter, every time the user clicks this shortcut, our payload will be launched with almost no perceptible difference.

### 高级效果 / Advanced Features
当不使用-i参数而使用-r参数的时候，程序自动替换指定目录下所有lnk，并且做了安全冗余设计，即使不同的快捷方式指向一个exe也不会报错，单线程的处理劫持不会使用户桌面打乱，而是维持原样，但是还是不建议在实战环境当中直接全部替换，一些杀软会检测自身状态，一旦修改则会报敏感，浏览器也是同样的道理，但是正常的第三方程序并无此效果

当使用-power参数后，vbs脚本会替换为强制管理员执行的版本，但由于执行vbs脚本的程序为win自带，所以并不会报未授权签名行为，和大部分的白程序uac认证界面相同，这可以做到强制无感要求用户以管理员上线我们的机器

但是需要注意，虽然这一切伪装的很好，但还是有聪明人会注意到为何一个简单的视频播放器启动需要管理员权限的，所以实际使用中请谨慎使用

When the `-r` parameter is used instead of `-i`, the program automatically replaces all shortcuts in the specified directory and incorporates safety redundancy design. Even if multiple shortcuts point to the same exe, no errors occur. The single-threaded hijacking process will not disrupt the user’s desktop but keep it intact.However, it is still not recommended to directly replace everything in a real-world environment. Some antivirus software can detect its own state, and any modifications may trigger sensitive alerts. The same principle applies to browsers, but regular third-party programs typically do not exhibit this behavior.

When the `-power` parameter is used, the vbs script is replaced by a version that forces administrator execution. Because the vbs script is executed by Windows built-in programs, there will be no unauthorized signature warnings. This behaves similarly to most legitimate programs’ UAC prompts, enabling stealthy forced elevation to run our payload as administrator.

However, it is important to note that although this is well disguised, some savvy users might notice why a Simple video player would require administrator privileges. Please use this feature with caution in practice.

# 使用视频 / Usage Video
还没时间录制。  
No video recording available yet.
# how to use
用法 / Usage:
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
# 免责声明
本工具仅供安全研究和教育用途。使用本工具进行任何未经授权的入侵、攻击或破坏行为均属违法行为，开发者不承担因此产生的任何法律责任。用户必须确保在合法授权范围内使用本工具，否则后果自负。
# Disclaimer
This tool is intended for security research and educational purposes only. Any unauthorized hacking, attacks, or damage caused by using this tool are illegal. The developer assumes no responsibility for any legal consequences resulting from misuse. Users must ensure the tool is used only within the scope of legal authorization, otherwise, all consequences are at the user's own risk.
