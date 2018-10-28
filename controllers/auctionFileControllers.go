package controllers

import (
	"../dao"
	u "../utils"
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var CreateAuctionFile = func(w http.ResponseWriter, r *http.Request) {
	buf := &bytes.Buffer{}
	buf, name, extension := readFile(r, buf)

	vars := mux.Vars(r)
	auctionId, _ := strconv.ParseUint(vars["id"], 10, 32)
	auctionFile := &dao.AuctionFile{Name: name, Extension: extension, AuctionID: uint(auctionId)}
	resp := auctionFile.Create(buf)
	buf.Reset()
	u.Respond(w, resp)
}

var GetAuctionFileById = func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auctionFileId, _ := strconv.ParseUint(vars["fileId"], 10, 32)

	data, auctionFile := dao.GetAuctionFile(uint(auctionFileId))

	fileName := fmt.Sprintf("%s.%s", auctionFile.Name, auctionFile.Extension)

	u.RespondWithFile(w, data, fileName)
}

func readFile(r *http.Request, buf *bytes.Buffer) (*bytes.Buffer, string, string) {
	file, header, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	name := strings.Split(header.Filename, ".")
	io.Copy(buf, file)
	return buf, name[0], name[1]
}