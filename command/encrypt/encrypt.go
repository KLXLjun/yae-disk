package encrypt

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

func Md5SumStr(file string) string {
	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer f.Close()
	r := bufio.NewReader(f)

	h := md5.New()
	_, err = io.Copy(h, r)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func Md5SumFile(file io.Reader) string {
	h := md5.New()
	_, err := io.Copy(h, file)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Crc32SumFile(file io.Reader) string {
	h := crc32.NewIEEE()
	_, err := io.Copy(h, file)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Crc32SumStr(file string) string {
	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer f.Close()
	r := bufio.NewReader(f)

	h := sha1.New()
	_, err = io.Copy(h, r)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func Crc32Str(str string) string {
	return fmt.Sprintf("%x", crc32.ChecksumIEEE([]byte(str)))
}

func Sha1SumFile(file io.Reader) string {
	h := sha1.New()
	_, err := io.Copy(h, file)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func Sha1SumStr(file string) string {
	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer f.Close()
	r := bufio.NewReader(f)

	h := sha1.New()
	_, err = io.Copy(h, r)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func Sha256SumFile(file io.Reader) string {
	h := sha256.New()
	_, err := io.Copy(h, file)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func Sha256SumStr(file string) string {
	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer f.Close()
	r := bufio.NewReader(f)

	h := sha256.New()
	_, err = io.Copy(h, r)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
