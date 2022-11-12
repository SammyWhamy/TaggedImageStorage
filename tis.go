package main

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

const TagCount = 4
const FileCount = 8
const TagListOffset = 12
const TagListLength = 16
const IndexOffset = 20
const IndexLength = 24
const FolderLength = 28
const FolderOffset = 32

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"

var verbose = false

func main() {
	rand.Seed(time.Now().UnixNano())

	if len(os.Args) < 2 {
		fmt.Println("No arguments given.")
		os.Exit(1)
	}

	args := os.Args[1:]

	switch args[0] {
	case "init":
		tisInit()
	case "help", "--help", "-h":
		tisHelp()
	case "version", "--version", "-v":
		fmt.Println("Tagged Image Storage v1.0.0")
		os.Exit(0)
	}

	exclusiveMode := false
	moveFile := true
	setFileName := ""

	excludeTags := ""
	contains, index := Contains(args, "--exclude")
	if contains {
		excludeTags = args[index+1]
		args = append(args[:index], args[index+2:]...)
	}

	for _, arg := range args {
		if arg == "--exclusive" {
			exclusiveMode = true
		} else if arg == "--no-move" {
			moveFile = false
		} else if arg == "--verbose" || arg == "-V" {
			verbose = true
		} else if strings.HasPrefix(arg, "--file-name=") {
			setFileName = strings.Split(arg, "=")[1]
			if setFileName == "*" {
				split := strings.Split(args[1], ".")
				setFileName = randomString(16) + "." + split[len(split)-1]
			}
		}
	}

	nonFlagArgs := make([]string, 0)
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			nonFlagArgs = append(nonFlagArgs, arg)
		}
	}

	printVerbose("Using options:")
	printVerbose("    Exclude tags:", excludeTags)
	printVerbose("    Exclusive mode:", exclusiveMode)
	printVerbose("    Move file:", moveFile)
	printVerbose("    Verbose:", verbose)
	printVerbose("    Set file name:", setFileName)
	printVerbose()

	fileHandle, err := getFileHandle(nil)

	if err != nil {
		errorAndExit("Error opening index file.", err, fileHandle)
	}

	switch nonFlagArgs[0] {
	case "info":
		tisInfo(fileHandle)
	case "add-file":
		{
			if len(nonFlagArgs) < 4 {
				errorAndExit("Not enough arguments.", nil, fileHandle)
			}
			tisAddFile(nonFlagArgs[1:3], setFileName, moveFile, fileHandle)
		}
	case "list":
		{
			if len(nonFlagArgs) < 2 {
				errorAndExit("Not enough arguments.", nil, fileHandle)
			}
			tisList(nonFlagArgs[1], excludeTags, exclusiveMode, fileHandle)
		}
	case "random":
		{
			if len(nonFlagArgs) > 1 {
				tisRandom(nonFlagArgs[1], excludeTags, fileHandle)
			} else {
				tisRandom("", excludeTags, fileHandle)
			}
		}
	default:
		tisUnknownCommand()
	}

	_ = fileHandle.Close()
}

func tisUnknownCommand() {
	fmt.Println("Unknown command.")
	fmt.Println("Use 'tis help' to see a list of commands.")
	os.Exit(1)
}

func tisHelp() {
	fmt.Println("Help for Tagged Image Storage")
	fmt.Println("Usage: tis <command> [arguments] [options]")
	fmt.Println("Basic info:")
	fmt.Println("    Options are specified after the command, in any order.")
	fmt.Println("    File lists and tags are surrounded by double quotes (\")")
	fmt.Println("    Tags are separated by semicolons (;)")
	fmt.Println("    Commands prefixed with !! are not implemented yet.")
	fmt.Println("Commands:")
	fmt.Println("    init - Initialize a new index file.")
	fmt.Println("    info - Show information about the index file.")
	fmt.Println("    add-file <file> <tags> - Add a file to the index.")
	fmt.Println("    !! remove-file <file> - Remove a file from the index.")
	fmt.Println("    list <tags> - List all files with the given tags.")
	fmt.Println("    random <?tags> - Get a random file from the index.")
	fmt.Println("    !! add-tag <file> <tag> - Add a tag to a file.")
	fmt.Println("    !! remove-tag <file> <tag> - Remove a tag from a file.")
	fmt.Println("    !! file-tags <file> - List all tags for a file.")
	fmt.Println("    !! delete-tag <tag> - Delete a tag from the index.")
	fmt.Println("    !! rename-tag <old> <new> - Rename a tag.")
	fmt.Println("    !! rename-data-folder <new> - Rename the data folder.")
	fmt.Println("    !! export <?filename> - Export the index to a human-readable file.")
	fmt.Println("    help - Show this.")
	fmt.Println("    version - Show the version.")
	fmt.Println("Flags:")
	fmt.Println("    --exclusive (list) - Only show files with all tags.")
	fmt.Println("    --exclude <tags> (list, random) - Exclude files with the given tags.")
	fmt.Println("    --no-move (add-file) - Don't move the file to the data folder.")
	fmt.Println("    --file-name=<name> (add-file) - Set the name of the file in the data folder, use * for a random name that keeps the extension.")
	fmt.Println("    --verbose - Show debug output.")

	os.Exit(0)
}

func getFileHandle(file *os.File) (*os.File, error) {
	if file == nil {
		return os.OpenFile("index.tis", os.O_RDWR, 0666)
	} else {
		return file, nil
	}
}

func tisRandom(tags string, excludeTags string, fileHandle *os.File) {
	tagString := ""

	if tags != "" {
		tagString = tags
		printVerbose("Using tags:", tagString)
	} else {
		tagList := getTags(fileHandle)
		tagString = ""

		for _, tag := range tagList {
			tagString += tag.name + ";"
		}

		tagString = tagString[:len(tagString)-1]
	}

	files := getFilesByTags(tagString, fileHandle)
	filesToExclude := getFilesByTags(excludeTags, fileHandle)

	printVerbose("Found", len(files), "files")
	printVerbose("Excluding", len(filesToExclude), "files")

	var filteredFiles []string

	for _, file := range files {
		included := false
		for _, excludeFile := range filesToExclude {
			if file == excludeFile {
				included = true
				break
			}
		}

		if !included {
			filteredFiles = append(filteredFiles, file)
		}
	}

	printVerbose("Found", len(filteredFiles), "files.")

	if len(filteredFiles) == 0 {
		fmt.Println("No files found.")
		os.Exit(1)
	}

	exists := make(map[string]bool)
	uniqueFiles := make([]string, 0)

	for _, file := range filteredFiles {
		if !exists[file] {
			uniqueFiles = append(uniqueFiles, file)
			exists[file] = true
		}
	}

	fmt.Println(uniqueFiles[rand.Intn(len(uniqueFiles))])
}

func tisList(tags string, excludeTags string, exclusiveMode bool, fileHandle *os.File) {
	files := getFilesByTags(tags, fileHandle)
	filesToExclude := getFilesByTags(excludeTags, fileHandle)

	for _, file := range filesToExclude {
		for i, f := range files {
			if f == file {
				files = append(files[:i], files[i+1:]...)
			}
		}
	}

	exclusiveFiles := make([]string, 0)

	if exclusiveMode {
		tagCount := len(strings.Split(tags, ";"))
		counts := make(map[string]int)

		for _, file := range files {
			counts[file]++
		}

		for file, count := range counts {
			if count == tagCount {
				exclusiveFiles = append(exclusiveFiles, file)
			}
		}
	}

	if exclusiveMode {
		fmt.Println(strings.Join(exclusiveFiles, ", "))
	} else {
		fmt.Println(strings.Join(files, ", "))
	}

}

func getFilesByTags(tags string, fileHandle *os.File) []string {
	splitTags := strings.Split(strings.Replace(tags, "\"", "", -1), ";")
	tagList := getTags(fileHandle)

	files := make([]string, 0)

	for _, tag := range splitTags {
		for _, tagListTag := range tagList {
			if tag == tagListTag.name {
				fileNameLength := readIntAtOffset(int64(tagListTag.offset), fileHandle)
				fileNameList := string(readBytes(tagListTag.offset+4, fileNameLength, fileHandle))

				re := regexp.MustCompile(`[a-zA-Z0-9-]*\.(jpg|png|gif|jpeg|webp)`)
				fileNames := re.FindAllString(fileNameList, -1)
				files = append(files, fileNames...)
			}
		}
	}

	return files
}

func tisAddFile(args []string, setFileName string, moveFile bool, fileHandle *os.File) {
	printVerbose("Adding file to index.")

	filePath := strings.Replace(args[0], "\"", "", -1)
	fileTags := strings.Split(strings.Replace(args[1], "\"", "", -1), ";")
	fileCount := readIntAtOffset(FileCount, fileHandle)
	dataFolder := getDataFolderName(fileHandle)

	err := error(nil)

	if moveFile {
		printVerbose("Using data folder:", dataFolder)

		_, err = os.Stat(dataFolder)
		if os.IsNotExist(err) {
			printVerbose("Data folder does not exist, creating it.")
			err = os.Mkdir(dataFolder, 0777)
			if err != nil {
				errorAndExit("Error creating data folder.", err, fileHandle)
			}
		} else if err != nil {
			errorAndExit("Error checking if data folder exists.", err, fileHandle)
		}
	}

	fileName := ""
	if setFileName == "" {
		fileName = filePath[strings.LastIndex(filePath, "/")+1:]
	} else {
		fileName = setFileName
	}

	if moveFile {
		_, err = os.Stat(dataFolder + "/" + fileName)
		if os.IsNotExist(err) {
			printVerbose("Moving file to data folder.")
			if setFileName != "" {
				printVerbose("Using file name:", fileName)
			}
			err = os.Rename(filePath, dataFolder+"/"+fileName)
			if err != nil {
				errorAndExit("Error moving file to data folder.", err, fileHandle)
			}
		} else if err != nil {
			errorAndExit("Error checking if file exists in data folder.", err, fileHandle)
		} else {
			errorAndExit("File already exists in data folder.", err, fileHandle)
		}
	} else if setFileName != "" {
		_, err = os.Stat(fileName)
		if os.IsNotExist(err) {
			printVerbose("Renaming file to:", fileName)
			err = os.Rename(filePath, fileName)
			if err != nil {
				errorAndExit("Error renaming file.", err, fileHandle)
			}
		} else if err != nil {
			errorAndExit("Error checking if file exists.", err, fileHandle)
		} else {
			errorAndExit("File already exists, cannot rename.", err, fileHandle)
		}
	}

	writeIntAtOffset(FileCount, fileCount+1, fileHandle)
	addTags(fileTags, fileHandle)
	addFile(fileName, fileTags, fileHandle)

	fmt.Println("Added file to index.")
}

func tisInfo(fileHandle *os.File) {
	fmt.Println("Index file info:")

	tagCount := readIntAtOffset(TagCount, fileHandle)
	fileCount := readIntAtOffset(FileCount, fileHandle)
	tags := getTags(fileHandle)
	dataFolder := getDataFolderName(fileHandle)

	fmt.Println("Data folder:", dataFolder)
	fmt.Println("Tag count:", tagCount)
	fmt.Println("File count:", fileCount)

	tagNames := make([]string, 0)
	for _, tag := range tags {
		tagNames = append(tagNames, tag.name)
	}
	fmt.Println("Tags:", strings.Join(tagNames, ", "))
}

func tisInit() {
	fmt.Println("Initializing index file.")

	file, err := os.OpenFile("index.tis", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)

	if err == os.ErrExist {
		errorAndExit("Index file already exists.", err, file)
	}

	if err != nil {
		errorAndExit("Error creating index file.", err, nil)
	}

	fileBytes := []byte{
		0x2E, 0x54, 0x49, 0x53, // File header
		0x00, 0x00, 0x00, 0x00, // Tag count
		0x00, 0x00, 0x00, 0x00, // File count
		0x00, 0x00, 0x00, 0x24, // Tag list offset
		0x00, 0x00, 0x00, 0x00, // Tag list length
		0x00, 0x00, 0x00, 0x00, // Index offset
		0x00, 0x00, 0x00, 0x00, // Index length
		0x00, 0x00, 0x00, 0x04, // Data folder name length
		0x64, 0x61, 0x74, 0x61, // Data folder name
	}

	_, _ = file.Write(fileBytes)
	_ = file.Close()

	fmt.Println("Index file created.")

	os.Exit(0)
}

func getDataFolderName(fileHandle *os.File) string {
	dataLength := readIntAtOffset(FolderLength, fileHandle)
	dataBytes := readBytes(FolderOffset, dataLength, fileHandle)
	return string(dataBytes)
}

func getTags(fileHandle *os.File) []Tag {
	tagCount := readIntAtOffset(TagCount, fileHandle)
	tagListOffset := readIntAtOffset(TagListOffset, fileHandle)

	_, _ = fileHandle.Seek(int64(tagListOffset), 0)

	tagList := make([]Tag, tagCount)

	for i := 0; i < int(tagCount); i++ {
		tagOffsetBytes := make([]byte, 4)
		_, _ = fileHandle.Read(tagOffsetBytes)
		tagList[i].offset = int32FromArray(tagOffsetBytes)

		tagLengthBytes := make([]byte, 4)
		_, _ = fileHandle.Read(tagLengthBytes)
		tagLength := int32FromArray(tagLengthBytes)

		tagBytes := make([]byte, tagLength)
		_, _ = fileHandle.Read(tagBytes)
		tagList[i].name = string(tagBytes)
	}

	return tagList
}

func addTags(tags []string, fileHandle *os.File) {
	tagList := getTags(fileHandle)
	tagListNames := make([]string, len(tagList))
	for i, tag := range tagList {
		tagListNames[i] = tag.name
	}

	tagCount := len(tagListNames)
	tagListBytes := make([]byte, 0)

	tagListOffset := readIntAtOffset(TagListOffset, fileHandle)
	tagListLength := readIntAtOffset(TagListLength, fileHandle)

	fileListOffset := tagListOffset + tagListLength

	for _, tag := range tags {
		contains, _ := Contains(tagListNames, tag)
		if !contains {
			fileListOffset += int32(8 + len(tag))
		}
	}

	for _, tag := range tags {
		contains, _ := Contains(tagListNames, tag)
		if !contains {
			tagCount++
			tagListBytes = append(tagListBytes, []byte{0x00, 0x00, 0x00, 0x00}...)
			tagListBytes = append(tagListBytes, arrayFromInt32(int32(len(tag)))...)
			tagListBytes = append(tagListBytes, []byte(tag)...)
		}
	}

	if len(tagListBytes) > 0 {
		indexLength := readIntAtOffset(IndexLength, fileHandle)

		tempBytes := make([]byte, 0)

		if indexLength > 0 {
			indexOffset := readIntAtOffset(IndexOffset, fileHandle)
			tempBytes = readBytes(indexOffset, indexLength, fileHandle)
		}

		existingTagListBytes := readBytes(tagListOffset, tagListLength, fileHandle)

		writeBytes(tagListOffset+tagListLength, tagListBytes, fileHandle)

		writeIntAtOffset(TagListLength, tagListLength+int32(len(tagListBytes)), fileHandle)
		writeIntAtOffset(TagCount, int32(tagCount), fileHandle)
		writeIntAtOffset(IndexOffset, tagListOffset+tagListLength+int32(len(tagListBytes)), fileHandle)

		for i := 0; i < len(existingTagListBytes); {
			currentTagOffset := int32FromArray(existingTagListBytes[i : i+4])
			writeIntAtOffset(int64(tagListOffset+int32(i)), currentTagOffset+int32(len(tagListBytes)), fileHandle)
			i += 8 + int(int32FromArray(existingTagListBytes[i+4:i+8]))
		}

		if len(tempBytes) > 0 {
			writeBytes(tagListOffset+tagListLength+int32(len(tagListBytes)), tempBytes, fileHandle)
		}
	}
}

func addFile(fileName string, tags []string, fileHandle *os.File) {
	for _, tag := range tags {
		tagList := getTags(fileHandle)

		for _, t := range tagList {
			if t.name == tag {
				indexLength := readIntAtOffset(IndexLength, fileHandle)

				if t.offset == 0 {
					indexOffset := readIntAtOffset(IndexOffset, fileHandle)
					t.offset = indexOffset + indexLength
				}

				fileListLength := readIntAtOffset(int64(t.offset), fileHandle)

				writeIntAtOffset(int64(t.offset), fileListLength+int32(len(fileName)), fileHandle)

				if fileListLength != 0 && indexLength-fileListLength > 0 {
					tempBytes := readBytes(t.offset+4+fileListLength, indexLength-fileListLength, fileHandle)
					writeBytes(t.offset+4+fileListLength+int32(len(fileName)), tempBytes, fileHandle)
				}

				writeBytes(t.offset+4+fileListLength, []byte(fileName), fileHandle)
				writeIntAtOffset(IndexLength, indexLength+int32(len(fileName))+4, fileHandle)

				tagListOffset := readIntAtOffset(TagListOffset, fileHandle)
				tagListLength := readIntAtOffset(TagListLength, fileHandle)
				tagListBytes := readBytes(tagListOffset, tagListLength, fileHandle)

				found := false

				for i := 0; i < len(tagListBytes); {
					currentTagOffset := int32FromArray(tagListBytes[i : i+4])
					tagLength := int32FromArray(tagListBytes[i+4 : i+8])
					tagName := string(tagListBytes[i+8 : i+8+int(tagLength)])

					if tagName != tag && !found {

					} else if !found {
						writeIntAtOffset(int64(int(tagListOffset)+i), t.offset, fileHandle)
						found = true
					} else if currentTagOffset != 0 {
						writeIntAtOffset(int64(int(tagListOffset)+i), currentTagOffset+int32(len(fileName)), fileHandle)
					}

					i += 8 + int(tagLength)
				}
			}
		}
	}
}

func readBytes(offset int32, length int32, fileHandle *os.File) []byte {
	_, _ = fileHandle.Seek(int64(offset), 0)
	bytes := make([]byte, length)
	_, _ = fileHandle.Read(bytes)
	return bytes
}

func writeBytes(offset int32, bytes []byte, fileHandle *os.File) {
	_, _ = fileHandle.Seek(int64(offset), 0)
	_, _ = fileHandle.Write(bytes)
}

func int32FromArray(bytes []byte) int32 {
	return int32(bytes[3]) | int32(bytes[2])<<8 | int32(bytes[1])<<16 | int32(bytes[0])<<24
}

func arrayFromInt32(i int32) []byte {
	return []byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)}
}

func readIntAtOffset(offset int64, fileHandle *os.File) int32 {
	fileBytes := make([]byte, 4)

	_, _ = fileHandle.Seek(offset, 0)
	_, _ = fileHandle.Read(fileBytes)

	data := int32FromArray(fileBytes)

	return data
}

func writeIntAtOffset(offset int64, data int32, fileHandle *os.File) {
	_, _ = fileHandle.Seek(offset, 0)
	_, _ = fileHandle.Write(arrayFromInt32(data))
}

func Contains[T comparable](s []T, e T) (bool, int) {
	for i, v := range s {
		if v == e {
			return true, i
		}
	}
	return false, -1
}

func errorAndExit(message string, err error, fileHandle *os.File) {
	if fileHandle != nil {
		_ = fileHandle.Close()
	}

	if err != nil {
		fmt.Println(message, "\n\t", err)
	} else {
		fmt.Println(message)
	}

	os.Exit(1)
}

func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Int63()%int64(len(letters))]
	}
	return string(b)
}

func printVerbose(a ...any) {
	if verbose {
		fmt.Println(a...)
	}
}

type Tag struct {
	name   string
	offset int32
}
