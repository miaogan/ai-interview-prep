// Package resume extracts plain text and structured profiles from resume files.
package resume

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/ledongthuc/pdf"
)

// maxExtractLen caps extracted text sent to the LLM.
const maxExtractLen = 20000

// ExtractText converts an uploaded resume file (pdf/docx/txt/md) to plain text.
// Unsupported formats and empty content are hard errors.
func ExtractText(filename string, data []byte) (string, error) {
	ext := strings.ToLower(path.Ext(filename))
	var text string
	var err error
	switch ext {
	case ".pdf":
		text, err = extractPDF(data)
	case ".docx":
		text, err = extractDocx(data)
	case ".txt", ".md":
		text = string(data)
	case ".doc":
		return "", errors.New("暂不支持 .doc 二进制格式，请另存为 .docx 或 .pdf")
	default:
		return "", fmt.Errorf("不支持的文件格式 %q，请上传 pdf / docx / md / txt", ext)
	}
	if err != nil {
		return "", err
	}
	text = strings.TrimSpace(text)
	if len([]rune(text)) < 20 {
		return "", fmt.Errorf("从 %s 中未能提取到有效文本（可能为扫描件或加密文件）", filename)
	}
	if runes := []rune(text); len(runes) > maxExtractLen {
		text = string(runes[:maxExtractLen])
	}
	return text, nil
}

func extractPDF(data []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("PDF 解析失败: %w", err)
	}
	pr, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("PDF 文本提取失败: %w", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, pr); err != nil {
		return "", fmt.Errorf("PDF 文本读取失败: %w", err)
	}
	return buf.String(), nil
}

// wordDocument is a minimal projection of word/document.xml.
type wordDocument struct {
	Paragraphs []struct {
		Runs []struct {
			Text string `xml:"t"`
		} `xml:"r"`
	} `xml:"body>p"`
}

func extractDocx(data []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("DOCX 解析失败（不是有效的 zip 容器）: %w", err)
	}
	for _, f := range zr.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("DOCX 读取失败: %w", err)
		}
		defer rc.Close()
		var doc wordDocument
		if err := xml.NewDecoder(rc).Decode(&doc); err != nil {
			return "", fmt.Errorf("DOCX 内容解析失败: %w", err)
		}
		var sb strings.Builder
		for _, p := range doc.Paragraphs {
			for _, r := range p.Runs {
				sb.WriteString(r.Text)
			}
			sb.WriteByte('\n')
		}
		return sb.String(), nil
	}
	return "", errors.New("DOCX 中未找到 word/document.xml")
}
