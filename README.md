<h1 align="center">SeekTune :musical_note:</h1>

<p align="center">
  <a href="https://drive.google.com/file/d/1I2esH2U4DtXHsNgYbUi4OL-ukV5i_1PI/view">
  <img src="https://github.com/user-attachments/assets/e4d01e9c-05cf-4f35-acbc-1e3cd79d1e00" 
       alt="screenshot" 
       width="500">
</a>
</p>

<p align="center"><a href="https://drive.google.com/file/d/1I2esH2U4DtXHsNgYbUi4OL-ukV5i_1PI/view">Demo in Video</a></p>

## Description 🎼

SeekTune is an implementation of Shazam's song recognition algorithm based on insights from these [resources](#resources--card_file_box). The original version integrates Spotify and YouTube APIs to find and download songs; this version is modified to specifically adapt to PhiZone's needs and only preserves the core algorithm, removing online features and support for recording input. Support for SQLite is also removed for some issues regarding version incompatibilities.

[//]: # (## Current Limitations
While the algorithm works excellently in matching a song with its exact file, it doesn't always find the right match from a recording. However, this project is still a work in progress. I'm hopeful about making it work, but I could definitely use some help :slightly_smiling_face:.  
Additionally, it currently only supports song files in WAV format.
)

## Installation :desktop_computer:

### Prerequisites

- Golang: [Install Golang](https://golang.org/dl/)
- FFmpeg: [Install FFmpeg](https://ffmpeg.org/download.html)

### Steps

Clone the repository:

```
git clone https://github.com/cgzirim/seek-tune.git
```

Install dependencies for the backend

```
cd seek-tune
go get ./...
```

Install dependencies for the client

```
cd seek-tune/client
npm install
```

## Usage :bicyclist:

#### ▸ Start the Client App 🏃‍♀️‍➡️

```
# Assuming you're in the client directory:

npm start
```

#### ▸ Start the Backend App 🏃‍♀️

In a separate terminal window:

```
cd seek-tune
go run *.go serve [-proto <http|https> (default: http)] [-port <port number> (default: 5000)]
```

#### ▸ Find matches for a song/recording 🔎

```
go run *.go find <path-to-wav-file>
```

#### ▸ Delete fingerprints and songs 🗑️

```
go run *.go erase
```

## Example :film_projector:

Find matches of a song

```
$ go run *.go find songs/Voilà\ -\ André\ Rieu.wav
Top 20 matches:
        - Voilà by André Rieu, score: 5390686.00
        - I Am a Child of God by One Voice Children's Choir, score: 2539.00
        - I Have A Dream by ABBA, score: 2428.00
        - SOS by ABBA, score: 2327.00
        - Sweet Dreams (Are Made of This) - Remastered by Eurythmics, score: 2213.00
        - The Winner Takes It All by ABBA, score: 2094.00
        - Sleigh Ride by One Voice Children's Choir, score: 2091.00
        - Believe by Cher, score: 2089.00
        - Knowing Me, Knowing You by ABBA, score: 1958.00
        - Gimme! Gimme! Gimme! (A Man After Midnight) by ABBA, score: 1941.00
        - Take A Chance On Me by ABBA, score: 1932.00
        - Don't Stop Me Now - Remastered 2011 by Queen, score: 1892.00
        - I Do, I Do, I Do, I Do, I Do by ABBA, score: 1853.00
        - Everywhere - 2017 Remaster by Fleetwood Mac, score: 1779.00
        - You Will Be Found by One Voice Children's Choir, score: 1664.00
        - J'Imagine by One Voice Children's Choir, score: 1658.00
        - When You Believe by One Voice Children's Choir, score: 1629.00
        - When Love Was Born by One Voice Children's Choir, score: 1484.00
        - Don't Stop Believin' (2022 Remaster) by Journey, score: 1465.00
        - Lay All Your Love On Me by ABBA, score: 1436.00

Search took: 856.386557ms

Final prediction: Voilà by André Rieu , score: 5390686.00
```

## Database 👯‍♀️

This application uses MongoDB as the only supported database.

1. [Install MongoDB](https://www.mongodb.com/docs/manual/installation/)
2. Configure MongoDB Connection:  
   To connect to your MongoDB instance, set the following environment variables:

   - `DB_TYPE`: Set this to "mongo" to indicate using MongoDB.
   - `DB_USER`: The username for your MongoDB database.
   - `DB_PASS`: The password for your MongoDB database.
   - `DB_NAME`: The name of the MongoDB database you want to use.
   - `DB_HOST`: The hostname or IP address of your MongoDB server.
   - `DB_PORT`: The port number on which your MongoDB server is listening.

   **Note:** The database connection URI is constructed using the environment variables.  
   If the `DB_USER` or `DB_PASS` environment variables are not set, it defaults to connecting to `mongodb://localhost:27017`.

## Resources :card_file_box:

- [How does Shazam work - Coding Geek](https://drive.google.com/file/d/1ahyCTXBAZiuni6RTzHzLoOwwfTRFaU-C/view) (main resource)
- [Song recognition using audio fingerprinting](https://hajim.rochester.edu/ece/sites/zduan/teaching/ece472/projects/2019/AudioFingerprinting.pdf)
- [How does Shazam work - Toptal](https://www.toptal.com/algorithms/shazam-it-music-processing-fingerprinting-and-recognition)
- [Creating Shazam in Java](https://www.royvanrijn.com/blog/2010/06/creating-shazam-in-java/)

## Author :black_nib:

- Chigozirim Igweamaka
  - Check out my other [GitHub](https://github.com/cgzirim) projects.
  - Connect with me on [LinkedIn](https://www.linkedin.com/in/chigozirim-igweamaka/).
  - Follow me on [Twitter](https://twitter.com/cgzirim).

## License :lock:

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
