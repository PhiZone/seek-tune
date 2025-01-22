<h1 align="center">SeekTune :musical_note:</h1>

<p align="center">
  <a href="https://drive.google.com/file/d/1I2esH2U4DtXHsNgYbUi4OL-ukV5i_1PI/view">
  <img src="https://github.com/user-attachments/assets/e4d01e9c-05cf-4f35-acbc-1e3cd79d1e00" 
       alt="screenshot" 
       width="500">
</a>
</p>

<p align="center"><a href="https://drive.google.com/file/d/1I2esH2U4DtXHsNgYbUi4OL-ukV5i_1PI/view">视频演示</a></p>

## 描述 🎼

SeekTune 是基于这些[资源](#resources--card_file_box)的见解实现的 Shazam 歌曲识别算法。原始版本集成了 Spotify 和 YouTube API 来查找和下载歌曲；此版本专门针对 PhiZone 的需求进行了修改，仅保留核心算法，移除了在线功能和录音输入支持。由于版本不兼容问题，还移除了对 SQLite 的支持。

[//]: # (## 当前限制
虽然该算法在匹配歌曲文件方面表现出色，但并不总是能从录音中找到正确的匹配。然而，该项目仍在进行中。我对使其工作充满希望，但我确实需要一些帮助 :slightly_smiling_face:。
此外，它目前仅支持 WAV 格式的歌曲文件。
)

## 安装 :desktop_computer:

### 先决条件

- Golang: [安装 Golang](https://golang.org/dl/)
- FFmpeg: [安装 FFmpeg](https://ffmpeg.org/download.html)

### 步骤

克隆仓库：

```
git clone https://github.com/cgzirim/seek-tune.git
```

安装后端依赖项

```
cd seek-tune
go get ./...
```

安装客户端依赖项

```
cd seek-tune/client
npm install
```

## 使用 :bicyclist:

#### ▸ 启动客户端应用 🏃‍♀️‍➡️

```
# 假设你在客户端目录：

npm start
```

#### ▸ 启动后端应用 🏃‍♀️

在另一个终端窗口中：

```
cd seek-tune
go run *.go serve [-proto <http|https> (默认: http)] [-port <端口号> (默认: 5000)]
```

#### ▸ 查找歌曲/录音的匹配项 🔎

```
go run *.go find <wav 文件路径>
```

#### ▸ 删除指纹和歌曲 🗑️

```
go run *.go erase
```

## 示例 :film_projector:

查找歌曲的匹配项

```
$ go run *.go find songs/Voilà\ -\ André\ Rieu.wav
前 20 个匹配项：
        - Voilà by André Rieu, 分数: 5390686.00
        - I Am a Child of God by One Voice Children's Choir, 分数: 2539.00
        - I Have A Dream by ABBA, 分数: 2428.00
        - SOS by ABBA, 分数: 2327.00
        - Sweet Dreams (Are Made of This) - Remastered by Eurythmics, 分数: 2213.00
        - The Winner Takes It All by ABBA, 分数: 2094.00
        - Sleigh Ride by One Voice Children's Choir, 分数: 2091.00
        - Believe by Cher, 分数: 2089.00
        - Knowing Me, Knowing You by ABBA, 分数: 1958.00
        - Gimme! Gimme! Gimme! (A Man After Midnight) by ABBA, 分数: 1941.00
        - Take A Chance On Me by ABBA, 分数: 1932.00
        - Don't Stop Me Now - Remastered 2011 by Queen, 分数: 1892.00
        - I Do, I Do, I Do, I Do, I Do by ABBA, 分数: 1853.00
        - Everywhere - 2017 Remaster by Fleetwood Mac, 分数: 1779.00
        - You Will Be Found by One Voice Children's Choir, 分数: 1664.00
        - J'Imagine by One Voice Children's Choir, 分数: 1658.00
        - When You Believe by One Voice Children's Choir, 分数: 1629.00
        - When Love Was Born by One Voice Children's Choir, 分数: 1484.00
        - Don't Stop Believin' (2022 Remaster) by Journey, 分数: 1465.00
        - Lay All Your Love On Me by ABBA, 分数: 1436.00

搜索耗时: 856.386557ms

最终预测: Voilà by André Rieu , 分数: 5390686.00
```

## 数据库 👯‍♀️

此应用程序使用 MongoDB 作为唯一支持的数据库。

1. [安装 MongoDB](https://www.mongodb.com/docs/manual/installation/)
2. 配置 MongoDB 连接：  
   要连接到你的 MongoDB 实例，请设置以下环境变量：

   - `DB_TYPE`: 设置为 "mongo" 以指示使用 MongoDB。
   - `DB_USER`: 你的 MongoDB 数据库的用户名。
   - `DB_PASS`: 你的 MongoDB 数据库的密码。
   - `DB_NAME`: 你想使用的 MongoDB 数据库的名称。
   - `DB_HOST`: 你的 MongoDB 服务器的主机名或 IP 地址。
   - `DB_PORT`: 你的 MongoDB 服务器正在监听的端口号。

   **注意:** 数据库连接 URI 是使用环境变量构建的。  
   如果未设置 `DB_USER` 或 `DB_PASS` 环境变量，则默认连接到 `mongodb://localhost:27017`。

## 资源 :card_file_box:

- [Shazam 如何工作 - Coding Geek](https://drive.google.com/file/d/1ahyCTXBAZiuni6RTzHzLoOwwfTRFaU-C/view) (主要资源)
- [使用音频指纹进行歌曲识别](https://hajim.rochester.edu/ece/sites/zduan/teaching/ece472/projects/2019/AudioFingerprinting.pdf)
- [Shazam 如何工作 - Toptal](https://www.toptal.com/algorithms/shazam-it-music-processing-fingerprinting-and-recognition)
- [用 Java 创建 Shazam](https://www.royvanrijn.com/blog/2010/06/creating-shazam-in-java/)

## 作者 :black_nib:

- Chigozirim Igweamaka
  - 查看我的其他 [GitHub](https://github.com/cgzirim) 项目。
  - 在 [LinkedIn](https://www.linkedin.com/in/chigozirim-igweamaka/) 上与我联系。
  - 在 [Twitter](https://twitter.com/cgzirim) 上关注我。

## 许可证 :lock:

此项目根据 MIT 许可证授权 - 有关详细信息，请参阅 [LICENSE](./LICENSE) 文件。
