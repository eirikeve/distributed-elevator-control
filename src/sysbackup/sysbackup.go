package sysbackup

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

/*

sysbackup has methods for making and recovering file backups of the elevator system.

*/

////////////////////////////////
// Module variables
////////////////////////////////

// Is set when when initialized, which happens automatically.
var initialized = false

// The file we log to, it is used by Backup.
var logFile *os.File

// Regexp used to identify backup files' names
var backupRegexp, _ = regexp.Compile("^backup_[0-9]+.elevlog$")

// Regexp used to identify the ID of a system in a backup entry
var idRegexp, _ = regexp.Compile("id=.+ backup=")

// Regexp used to identify the state of a system in a backup entry
var stateRegexp, _ = regexp.Compile("backup={.+}\n$")

// Where backup files are stored
const folderDir = "../backup/"

////////////////////////////////
// Interface
////////////////////////////////

/*Backup stores the passed argument states in the file logFile.
 * If logFile has not been set (if not initialized), it is created first.
 * @arg states: Slice containing states of systems.
 */
func Backup(states []et.ElevState) {
	if !initialized {
		setupSysBackup()
	}
	log.WithField("FileName", (*logFile).Name()).Debug("sysbackup Backup: Backed up")
	for i := 0; i < len(states); i++ {
		backupElevState(states[i])
	}

}

/*Recover searchs the ./backup folder for .elevlog backups, recovers them if sufficiently recent, and returns them
 * @arg timeLimit: Files must be modified after this to be recovered.
 * @return: Recovered backup. Nil if none found, or they were from before timeLimit.
 *			Err nonzero if we failed to scan for files.
 */
func Recover(timeLimit time.Time) ([]et.ElevState, error) {

	// Locate all files in the dir
	files, err := ioutil.ReadDir(folderDir)
	if err != nil {
		log.WithField("err", err.Error()).Error("sysbackup Recover: Failed")
		return make([]et.ElevState, 0), err
	}
	// We are only interested in .elevlog files
	backupFilesIndexes := getBackupFileIndexes(files)

	// Discard files that are from before timeLimit
	useableBackupIndexes := make([]int, 0)
	for _, index := range backupFilesIndexes {
		if files[index].ModTime().Sub(timeLimit) > 0*time.Second {
			useableBackupIndexes = append(useableBackupIndexes, index)
			log.WithField("usable file", files[index].Name()).Info("sysbackup Recover: Valid file for recovery")
		}
	}
	// Sort the files by time last modified
	sortedIndexes, _ := getBackupFileIndexesSortedInIncreasingTime(files, useableBackupIndexes)

	// Recover info from the backups
	states := make([]et.ElevState, 0)
	for _, backupIndex := range sortedIndexes {
		log.WithField("File", files[backupIndex].Name()).Info("sysbackup: Applying backup")
		applyBackupFromFile(&states, files[backupIndex])
	}
	numBackupFiles := strconv.FormatInt(int64(len(backupFilesIndexes)), 10)
	numUseableBackupFiles := strconv.FormatInt(int64(len(useableBackupIndexes)), 10)
	numStatesRecovered := strconv.FormatInt(int64(len(states)), 10)

	log.WithFields(log.Fields{"numStatesRecovered": numStatesRecovered,
		"numBackupFiles":        numBackupFiles,
		"numUseableBackupFiles": numUseableBackupFiles,
	}).Info("sysbackup Recover:")

	return states, nil
}

////////////////////////////////
// Auxiliary functions
////////////////////////////////

/*setupSysBackup sets the Backup output file.
 * If necessary, it also creates the ./backup directory
 */
func setupSysBackup() {

	filename := folderDir + "backup_" + strconv.FormatInt(time.Now().Unix(), 10) + ".elevlog"
	var err error
	logFile, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		// Could not open file
		os.MkdirAll(folderDir, os.ModePerm)
		log.WithError(err).Error("sysbackup Setup: Could not open log output file. Created backup folder " + folderDir)

		logFile, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.WithError(err).Error("sysbackup Setup: Could not initialize")
			initialized = false
		} else {
			initialized = true

		}

	} else {
		initialized = true

	}
}

/*getbackupFileIndexes determines which files among a set are .elevlog files.
 * @arg files: Slice of files in a folder
 * @return: Indexes of the files arg which are .elevlog files
 */
func getBackupFileIndexes(files []os.FileInfo) []int {
	backupFilesIndexes := make([]int, 0)
	for i, f := range files {
		if backupRegexp.MatchString(f.Name()) {
			backupFilesIndexes = append(backupFilesIndexes, i)
		}
	}
	return backupFilesIndexes
}

/*getBackupFileIndexesSortedInIncreasingTime returns the indexes of the backup files, sorted by when they were last modified.
 * Oldest are first, most recent last in the return slice
 * @arg files: Slice of files in a folder
 * @arg backupIndexes: Indexes of the files arg which are .elevlog files
 * @return: backupIndexes sorted by modification time
 */
func getBackupFileIndexesSortedInIncreasingTime(files []os.FileInfo, backupIndexes []int) ([]int, error) {
	timestamps := make([]int64, len(backupIndexes))
	sortedIndexes := make([]int, len(backupIndexes))
	for i, backupIndex := range backupIndexes {
		timestampAsString := strings.TrimSuffix(strings.TrimPrefix(files[backupIndex].Name(), "backup_"), ".elevlog")
		var err error
		timestamps[i], err = strconv.ParseInt(timestampAsString, 10, 64)
		if err != nil {
			return make([]int, 0), err
		}
	}
	for i, _ := range sortedIndexes {
		minimumIndex := 0
		for j, val := range timestamps {
			if val < timestamps[minimumIndex] {
				minimumIndex = j
			}
		}
		sortedIndexes[i] = backupIndexes[minimumIndex]

		if minimumIndex+1 < len(backupIndexes) {
			backupIndexes = append(backupIndexes[:minimumIndex], backupIndexes[minimumIndex+1:]...)
			timestamps = append(timestamps[:minimumIndex], timestamps[minimumIndex+1:]...)
		} else {
			backupIndexes = backupIndexes[:minimumIndex]
			timestamps = timestamps[:minimumIndex]
		}

	}
	return sortedIndexes, nil
}

/*applyBackupFromFile goes through a file, converts each line to a backuped object, and applies it to the states arg system.
 * @arg states: slice of system's states
 * @arg backupFile: A single .elevlog file
 */
func applyBackupFromFile(states *[]et.ElevState, backupFile os.FileInfo) {
	log.WithField("filaname", backupFile.Name()).Debug("sysbackup apply: Applying backup")
	file, err := os.OpenFile(folderDir+backupFile.Name(), os.O_RDONLY, 0755)
	if err != nil {
		log.WithError(err).Error("sysbackup: Unable to apply backup from this file")
		return
	}
	var backupReader = bufio.NewReader(file)
	for {
		line, err := backupReader.ReadString('\n')
		if err != nil {
			return
		}
		elevatorJson := getStateJSONFromBackup(&line)
		var state et.ElevState
		jsonErr := json.Unmarshal([]byte(elevatorJson), &state)
		if jsonErr != nil {
			log.WithField("err", jsonErr.Error()).Warn("sysbackup apply: Error applying backup")
		}
		if state.ID > 0 {
			if len(*states) == 0 {
				*states = append(*states, state)
			} else {
				for i, s := range *states {
					if s.ID == state.ID {
						(*states)[i] = state
						break // for i, s := range *states
					} else if i == len(*states)-1 {
						*states = append(*states, state)
						break // for i, s := range *states
					}
				}
			}
		}
	}
}

/*getStateJSONFromBackup finds the part of a line which represents a JSON-formatted et.ElevState object, and returns that part.
 * @arg line: Line from backup file
 * @return: Part of the line
 */
func getStateJSONFromBackup(line *string) string {
	stateWithSuffixAndPrefix := stateRegexp.FindString(*line)
	state := strings.TrimPrefix(strings.TrimSuffix(stateWithSuffixAndPrefix, "\n"), "backup=")
	return state
}

/*backupElevState stores a state as a JSON-formatted string in a backup file.
 * @arg state: State of a system
 */
func backupElevState(state et.ElevState) {
	if logFile == nil {
		log.WithField("err", "Nonexistent logfile").Error("sysbackup backup: Could not back up")
		return
	}
	buf, _ := json.Marshal(state)
	backup := string(buf)
	backupMsg := "time=" + strconv.FormatInt(time.Now().Unix(), 10) + " id=" + strconv.FormatInt(int64(state.ID), 10) + " backup=" + backup + "\n"

	logFile.Write([]byte(backupMsg))
}
