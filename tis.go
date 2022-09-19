package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No arguments given.")
		os.Exit(1)
	}

	args := os.Args[1:]

	if args[0] == "init" {
		tisInit()
		os.Exit(0)
	}

	fileHandle, err := getFileHandle(nil)

	if err != nil {
		errorAndExit("Error opening index file.", err, nil)
	}

	switch args[0] {
	case "info":
		tisInfo(fileHandle)
	case "add-file":
		tisAddFile(args[1:3], fileHandle)
	case "list":
		tisList(args[1], fileHandle)
	}

	_ = fileHandle.Close()
}

func getFileHandle(file *os.File) (*os.File, error) {
	if file == nil {
		return os.OpenFile("index.tis", os.O_RDWR, 0666)
	} else {
		return file, nil
	}
}

func tisList(tags string, fileHandle *os.File) {
	splitTags := strings.Split(strings.Replace(tags, "\"", "", -1), ";")

	tagList := getTags(fileHandle)

	files := make([]string, 0)

	for _, tag := range splitTags {
		for _, tagListTag := range tagList {
			if tag == tagListTag.name {
				fileNameLength := readIntAtOffset(int64(tagListTag.offset), fileHandle)

				fileNameBytes := make([]byte, fileNameLength)
				_, _ = fileHandle.Read(fileNameBytes)

				re := regexp.MustCompile(`[0-9]*\.(png|jpg|gif)`)
				fileNames := re.FindAllString(string(fileNameBytes), -1)
				files = append(files, fileNames...)
			}
		}
	}

	fmt.Println("File location:", "./"+getDataFolderName(fileHandle)+"/")
	fmt.Println("Files:", strings.Join(files, ", "))
}

func tisAddFile(args []string, fileHandle *os.File) {
	fmt.Println("Adding file to index.")

	fileName := strings.Replace(args[0], "\"", "", -1)
	fileExtension := fileName[strings.LastIndex(fileName, ".")+1:]
	fileTags := strings.Split(strings.Replace(args[1], "\"", "", -1), ";")
	fileCount := readIntAtOffset(8, fileHandle)
	dataFolder := getDataFolderName(fileHandle)

	fmt.Println("Using data folder:", dataFolder)

	_, err := os.Stat(dataFolder)
	if os.IsNotExist(err) {
		fmt.Println("Data folder does not exist, creating it.")
		err := os.Mkdir(dataFolder, 0777)
		if err != nil {
			errorAndExit("Error creating data folder.", err, fileHandle)
		}
	} else if err != nil {
		errorAndExit("Error checking if data folder exists.", err, fileHandle)
	}

	newFileName := strconv.Itoa(int(fileCount+1)) + "." + fileExtension

	_, err = os.Stat(dataFolder + "/" + fileName)
	if os.IsNotExist(err) {
		fmt.Println("Moving file to data folder.")
		err := os.Rename(fileName, dataFolder+"/"+newFileName)
		if err != nil {
			errorAndExit("Error moving file to data folder.", err, fileHandle)
		}
	} else if err != nil {
		errorAndExit("Error checking if file exists in data folder.", err, fileHandle)
	} else {
		fmt.Println("File already exists, aborting.")
		os.Exit(1)
	}

	writeIntAtOffset(8, fileCount+1, fileHandle)
	addTags(fileTags, fileHandle)
	addFile(newFileName, fileTags, fileHandle)
}

func tisInfo(fileHandle *os.File) {
	fmt.Println("Index file info:")

	tagCount := readIntAtOffset(4, fileHandle)
	fileCount := readIntAtOffset(8, fileHandle)
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
		fmt.Println("Index file already exists, aborting.")
		os.Exit(1)
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
}

func getDataFolderName(fileHandle *os.File) string {
	dataLength := readIntAtOffset(28, fileHandle)
	dataBytes := make([]byte, dataLength)

	_, _ = fileHandle.Seek(32, 0)
	_, _ = fileHandle.Read(dataBytes)

	return string(dataBytes)
}

func getTags(fileHandle *os.File) []Tag {
	tagCount := readIntAtOffset(4, fileHandle)
	tagListOffset := readIntAtOffset(12, fileHandle)

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

	tagListOffset := readIntAtOffset(12, fileHandle)
	tagListLength := readIntAtOffset(16, fileHandle)

	fileListOffset := tagListOffset + tagListLength

	for _, tag := range tags {
		if !Contains(tagListNames, tag) {
			fileListOffset += int32(8 + len(tag))
		}
	}

	for _, tag := range tags {
		if !Contains(tagListNames, tag) {
			tagCount++
			tagListBytes = append(tagListBytes, []byte{0x00, 0x00, 0x00, 0x00}...)
			tagListBytes = append(tagListBytes, arrayFromInt32(int32(len(tag)))...)
			tagListBytes = append(tagListBytes, []byte(tag)...)
		}
	}

	if len(tagListBytes) > 0 {
		indexLength := readIntAtOffset(24, fileHandle)
		tempBytes := make([]byte, indexLength)

		if indexLength > 0 {
			indexOffset := readIntAtOffset(20, fileHandle)

			_, _ = fileHandle.Seek(int64(indexOffset), 0)
			_, _ = fileHandle.Read(tempBytes)
		}

		existingTagListBytes := make([]byte, tagListLength)
		_, _ = fileHandle.Seek(int64(tagListOffset), 0)
		_, _ = fileHandle.Read(existingTagListBytes)

		_, _ = fileHandle.Seek(int64(tagListOffset+tagListLength), 0)
		_, _ = fileHandle.Write(tagListBytes)

		writeIntAtOffset(16, tagListLength+int32(len(tagListBytes)), fileHandle)
		writeIntAtOffset(4, int32(tagCount), fileHandle)
		writeIntAtOffset(20, tagListOffset+tagListLength+int32(len(tagListBytes)), fileHandle)

		for i := 0; i < len(existingTagListBytes); {
			currentTagOffset := int32FromArray(existingTagListBytes[i : i+4])
			writeIntAtOffset(int64(tagListOffset+int32(i)), currentTagOffset+int32(len(tagListBytes)), fileHandle)
			i += 8 + int(int32FromArray(existingTagListBytes[i+4:i+8]))
		}

		if len(tempBytes) > 0 {
			_, _ = fileHandle.Seek(int64(tagListOffset+tagListLength+int32(len(tagListBytes))), 0)
			_, _ = fileHandle.Write(tempBytes)
		}
	}
}

func addFile(fileName string, tags []string, fileHandle *os.File) {
	for _, tag := range tags {
		tagList := getTags(fileHandle)

		for _, t := range tagList {
			if t.name == tag {
				fmt.Println("Adding file: " + fileName + " to tag: " + tag)

				indexLength := readIntAtOffset(24, fileHandle)

				fmt.Println("Tag offset: " + strconv.Itoa(int(t.offset)))
				if t.offset == 0 {
					indexOffset := readIntAtOffset(20, fileHandle)

					t.offset = indexOffset + indexLength
					fmt.Println("New tag offset: " + strconv.Itoa(int(t.offset)))
				}

				fileListLength := readIntAtOffset(int64(t.offset), fileHandle)

				fmt.Println("File list length: " + strconv.Itoa(int(fileListLength)))

				writeIntAtOffset(int64(t.offset), fileListLength+int32(len(fileName)), fileHandle)

				_, _ = fileHandle.Seek(int64(t.offset+4+fileListLength), 0)

				tempBytesLength := 0
				if fileListLength != 0 {
					tempBytesLength = int(indexLength - fileListLength)
				}

				fmt.Println("Temp bytes length: " + strconv.Itoa(tempBytesLength))

				tempBytes := make([]byte, tempBytesLength)
				if len(tempBytes) > 0 {
					_, _ = fileHandle.Read(tempBytes)
				}

				fmt.Println("Writing filename at offset: " + strconv.Itoa(int(t.offset+fileListLength)))

				_, _ = fileHandle.Seek(int64(t.offset+4+fileListLength), 0)
				_, _ = fileHandle.Write([]byte(fileName))

				if len(tempBytes) > 0 {
					_, _ = fileHandle.Write(tempBytes)
				}

				writeIntAtOffset(24, indexLength+int32(len(fileName))+4, fileHandle)

				tagListOffset := readIntAtOffset(12, fileHandle)
				tagListLength := readIntAtOffset(16, fileHandle)

				_, _ = fileHandle.Seek(int64(tagListOffset), 0)
				tagListBytes := make([]byte, tagListLength)
				_, _ = fileHandle.Read(tagListBytes)

				fmt.Println("Tag list bytes: " + string(tagListBytes))

				found := false

				for i := 0; i < len(tagListBytes); {
					currentTagOffset := int32FromArray(tagListBytes[i : i+4])
					tagLength := int32FromArray(tagListBytes[i+4 : i+8])

					fmt.Println("Tag length: " + strconv.Itoa(int(tagLength)))

					tagName := string(tagListBytes[i+8 : i+8+int(tagLength)])

					fmt.Println("Updating offset for tag: " + tagName)

					if tagName != tag && !found {
						fmt.Println("Skipping.")

						i += 8 + int(tagLength)
						continue
					}

					if !found {
						fmt.Println("Updating current tag with new offset: ", t.offset)

						writeIntAtOffset(int64(int(tagListOffset)+i), t.offset, fileHandle)
						i += 8 + int(tagLength)
						found = true
						continue
					}

					if currentTagOffset != 0 {
						fmt.Println("Updating next tag with new offset: ", currentTagOffset+int32(len(fileName)))
						writeIntAtOffset(int64(int(tagListOffset)+i), currentTagOffset+int32(len(fileName)), fileHandle)
					}

					i += 8 + int(tagLength)
				}
			}
		}
	}
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

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func errorAndExit(message string, err error, fileHandle *os.File) {
	if fileHandle != nil {
		_ = fileHandle.Close()
	}
	fmt.Println(message, "\n\t", err)
	os.Exit(1)
}

type Tag struct {
	name   string
	offset int32
}
