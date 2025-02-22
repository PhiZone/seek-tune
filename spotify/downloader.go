package spotify

import (
	"fmt"
	"song-recognition/db"
	"song-recognition/shazam"
	"song-recognition/wav"

	"github.com/fatih/color"
)

const DELETE_SONG_FILE = true

var yellow = color.New(color.FgYellow)

func ProcessAndSaveSong(songFilePath, songTitle, songArtist, pzID string) error {
	dbclient, err := db.NewDBClient()
	if err != nil {
		return err
	}
	defer dbclient.Close()

	wavFilePath, err := wav.ConvertToWAV(songFilePath, 1)
	if err != nil {
		return err
	}

	wavInfo, err := wav.ReadWavInfo(wavFilePath)
	if err != nil {
		return err
	}

	samples, err := wav.WavBytesToSamples(wavInfo.Data)
	if err != nil {
		return fmt.Errorf("error converting wav bytes to float64: %v", err)
	}

	spectro, err := shazam.Spectrogram(samples, wavInfo.SampleRate)
	if err != nil {
		return fmt.Errorf("error creating spectrogram: %v", err)
	}

	songId, err := dbclient.RegisterSong(songTitle, songArtist, pzID)
	if err != nil {
		return err
	}

	peaks := shazam.ExtractPeaks(spectro, wavInfo.Duration)
	fingerprints := shazam.Fingerprint(peaks, songId)

	err = dbclient.StoreFingerprints(fingerprints)
	if err != nil {
		dbclient.DeleteSongByID(songId)
		return fmt.Errorf("error to storing fingerpring: %v", err)
	}

	fmt.Printf("Fingerprint for %v by %v saved in DB successfully\n", songTitle, songArtist)
	return nil
}

func ProcessAndUpdateSong(songFilePath, songTitle, songArtist, pzID string) error {
	dbclient, err := db.NewDBClient()
	if err != nil {
		return err
	}
	defer dbclient.Close()

	wavFilePath, err := wav.ConvertToWAV(songFilePath, 1)
	if err != nil {
		return err
	}

	wavInfo, err := wav.ReadWavInfo(wavFilePath)
	if err != nil {
		return err
	}

	samples, err := wav.WavBytesToSamples(wavInfo.Data)
	if err != nil {
		return fmt.Errorf("error converting wav bytes to float64: %v", err)
	}

	spectro, err := shazam.Spectrogram(samples, wavInfo.SampleRate)
	if err != nil {
		return fmt.Errorf("error creating spectrogram: %v", err)
	}

	err = dbclient.DeleteSongByPhiZoneID(pzID)
	if err != nil {
		return fmt.Errorf("error deleting song by PhiZoneID: %v", err)
	}

	songId, err := dbclient.RegisterSong(songTitle, songArtist, pzID)
	if err != nil {
		return err
	}

	peaks := shazam.ExtractPeaks(spectro, wavInfo.Duration)
	fingerprints := shazam.Fingerprint(peaks, songId)

	err = dbclient.StoreFingerprints(fingerprints)
	if err != nil {
		dbclient.DeleteSongByID(songId)
		return fmt.Errorf("error to storing fingerpring: %v", err)
	}

	fmt.Printf("Fingerprint for %v by %v saved in DB successfully\n", songTitle, songArtist)
	return nil
}

func ProcessAndSaveCopyrightSong(songFilePath, songTitle, songArtist, pzID string) error {
	dbclient, err := db.NewDBClient()
	if err != nil {
		return err
	}
	defer dbclient.Close()

	wavFilePath, err := wav.ConvertToWAV(songFilePath, 1)
	if err != nil {
		return err
	}

	wavInfo, err := wav.ReadWavInfo(wavFilePath)
	if err != nil {
		return err
	}

	samples, err := wav.WavBytesToSamples(wavInfo.Data)
	if err != nil {
		return fmt.Errorf("error converting wav bytes to float64: %v", err)
	}

	spectro, err := shazam.Spectrogram(samples, wavInfo.SampleRate)
	if err != nil {
		return fmt.Errorf("error creating spectrogram: %v", err)
	}

	PhiZoneID, err := dbclient.RegisterCopyrightSong(songTitle, songArtist, pzID)
	if err != nil {
		return err
	}

	peaks := shazam.ExtractPeaks(spectro, wavInfo.Duration)
	fingerprints := shazam.Fingerprint(peaks, PhiZoneID)

	err = dbclient.StoreFingerprints(fingerprints)
	if err != nil {
		dbclient.DeleteCopyrightSongByID(PhiZoneID)
		return fmt.Errorf("error to storing fingerpring: %v", err)
	}

	fmt.Printf("Fingerprint for %v by %v saved in DB successfully\n", songTitle, songArtist)
	return nil
}

func ProcessAndUpdateCopyrightSong(songFilePath, songTitle, songArtist, pzID string) error {
	dbclient, err := db.NewDBClient()
	if err != nil {
		return err
	}
	defer dbclient.Close()

	wavFilePath, err := wav.ConvertToWAV(songFilePath, 1)
	if err != nil {
		return err
	}

	wavInfo, err := wav.ReadWavInfo(wavFilePath)
	if err != nil {
		return err
	}

	samples, err := wav.WavBytesToSamples(wavInfo.Data)
	if err != nil {
		return fmt.Errorf("error converting wav bytes to float64: %v", err)
	}

	spectro, err := shazam.Spectrogram(samples, wavInfo.SampleRate)
	if err != nil {
		return fmt.Errorf("error creating spectrogram: %v", err)
	}

	err = dbclient.DeleteCopyrightSongByPhiZoneID(pzID)
	if err != nil {
		return fmt.Errorf("error deleting song by PhiZoneID: %v", err)
	}

	songId, err := dbclient.RegisterCopyrightSong(songTitle, songArtist, pzID)
	if err != nil {
		return err
	}

	peaks := shazam.ExtractPeaks(spectro, wavInfo.Duration)
	fingerprints := shazam.Fingerprint(peaks, songId)

	err = dbclient.StoreFingerprints(fingerprints)
	if err != nil {
		dbclient.DeleteCopyrightSongByID(songId)
		return fmt.Errorf("error to storing fingerpring: %v", err)
	}

	fmt.Printf("Fingerprint for %v by %v saved in DB successfully\n", songTitle, songArtist)
	return nil
}
