package thumbnailer

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/gif"	
	_ "image/png"
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
		
	"github.com/nfnt/resize"
	"github.com/3d0c/gmf"	
	
	"github.com/veriak/minion/config"
)

var (
	cfg		*config.Config
	minioClient	*minio.Client
	logger		= log.With().Str("service", "Minion").Logger()
)

func Start() {
	cfg = config.Get()	
	initMinio()	
	go doIt()
}

func initMinio() {
	var err error
	minioClient, err = minio.New(cfg.Minio.Addr, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKeyID, cfg.Minio.SecretAccessKey, ""),
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		panic(err)
	}
}
				
func doIt() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for notifInfo := range minioClient.ListenNotification(ctx, "", "", []string{
		//"s3:BucketCreated:*",
		//"s3:BucketRemoved:*",
		"s3:ObjectCreated:*",
		//"s3:ObjectAccessed:*",
		//"s3:ObjectRemoved:*",
	}) {

		if notifInfo.Err != nil {
			logger.Err(notifInfo.Err).Msgf("%v", notifInfo)
			return
		}
		
		event := notifInfo.Records[0]
		//logger.Info().Msgf("#### %v", event)
		
		bucketName := event.S3.Bucket.Name
		if bucketName != "image" && bucketName != "video" {
			continue	
		}
		
		objectName := event.S3.Object.Key						
		bucketName = "thumbnail"

		err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
			if errBucketExists == nil && exists {
				//logger.Info().Msgf("We already own %s\n", bucketName)
			} else {
				logger.Error().Err(err).Msg("minioClient.BucketExists")
				return
			}
		} else {
			logger.Info().Msgf("Successfully created %s", bucketName)
		}

		if checkFileExist(bucketName, objectName) {
			logger.Info().Msgf("Thumbnail already exist %s", objectName)
			continue
		}

		bucketName = event.S3.Bucket.Name		
		if bucketName == "image" || bucketName == "video" {
			reader, err := minioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
			if err != nil {
				logger.Error().Err(err).Msg("minioClient.GetObject")
				return
			}
			defer reader.Close()
			
			var buf bytes.Buffer			
			var thumbnail image.Image
			
			if bucketName == "image" {
				img, _, _ := image.Decode(reader)
				thumbnail = resize.Thumbnail(uint(cfg.Thumbnailer.Width), uint(cfg.Thumbnailer.Height), img, resize.Lanczos3)
			} else if bucketName == "video" {
			
				data, err := ioutil.ReadAll(reader)
				if err != nil {
					logger.Error().Err(err).Msg("Can't read resource " + objectName)
					return
				}
				//dataReader := bytes.NewReader(data)
				fileName := "/tmp/" + objectName
				err = ioutil.WriteFile(fileName, data, 0644)
				if err != nil {
					logger.Fatal().Err(err)
				}
    
				dataSize := int64(len(data))
				
				//inputCtx := gmf.NewCtx()
				inputCtx, err := gmf.NewInputCtx(fileName)
				if err != nil {
					logger.Fatal().Err(err)
				}    
				defer inputCtx.Free()
				
				logger.Info().Msgf("AVIOContext Duration: %d", int(inputCtx.Duration()))
				
				/*					
				var section *io.SectionReader
				
				avioCtx, err := gmf.NewAVIOContext(inputCtx, &gmf.AVIOHandlers{ReadPacket: func ()([]byte, int) {
					if section == nil {					
						section = io.NewSectionReader(dataReader, 0, dataSize)
					}
	
					b := make([]byte, gmf.IO_BUFFER_SIZE)
					n, err := section.Read(b)
					if err != nil && err == io.EOF {						
					}
					if err != nil {
						logger.Error().Err(err).Msg("section.Read()")
					}
					return b, n
				}})				
				if err != nil {
					logger.Fatal().Err(err).Msg("NewAVIOContext")
				}
				defer avioCtx.Free()
								
				if err = inputCtx.SetPb(avioCtx).OpenInput(""); err != nil {
					logger.Fatal().Err(err).Msg("AVIOContext OpenInput")
				}*/				
				
				srcVideoStream, err := inputCtx.GetBestStream(gmf.AVMEDIA_TYPE_VIDEO)
				if err != nil {
					logger.Fatal().Err(err).Msg("No video stream found")
				}				
				
				codec, err := gmf.FindEncoder(gmf.AV_CODEC_ID_RAWVIDEO/*AV_CODEC_ID_MJPEG*/)
				if err != nil {
					logger.Fatal().Err(err).Msg("FindEncoder")
				}
				
				cc := gmf.NewCodecCtx(codec)
				defer gmf.Release(cc)

				cc.SetPixFmt(gmf.AV_PIX_FMT_RGBA/*AV_PIX_FMT_YUVJ420P*/).
					SetWidth(srcVideoStream.CodecCtx().Width()).
					SetHeight(srcVideoStream.CodecCtx().Height()).
					SetTimeBase(gmf.AVR{Num: 1, Den: 1})

				if codec.IsExperimental() {
					cc.SetStrictCompliance(gmf.FF_COMPLIANCE_EXPERIMENTAL)
				}

				if err := cc.Open(nil); err != nil {
					logger.Fatal().Err(err)
				}
				defer cc.Free()

				ist, err := inputCtx.GetStream(srcVideoStream.Index())
				if err != nil {
					logger.Fatal().Err(err).Msg("Error getting stream")
				}
				defer ist.Free()
	
				icc := srcVideoStream.CodecCtx()
				
				logger.Info().Msgf("%v %v %v %v %v %v %v", dataSize, icc.Width(), icc.Height(), icc.PixFmt(), cc.Width(), cc.Height(), cc.PixFmt())
				
				swsCtx, err := gmf.NewSwsCtx(icc.Width(), icc.Height(), icc.PixFmt(), cc.Width(), cc.Height(), cc.PixFmt(), gmf.SWS_BICUBIC)
				if err != nil {
					logger.Fatal().Err(err)
				}
				defer swsCtx.Free()

				var (
					pkt        *gmf.Packet
					frames     []*gmf.Frame
					drain      int = -1
					frameCount int = 0
				)				
				
				bContinue := true

				for bContinue {					
					if drain >= 0 {
						break
					}					
					
					pkt, err = inputCtx.GetNextPacket()
					if err != nil && err != io.EOF {
						if pkt != nil {
							pkt.Free()
						}
						logger.Error().Err(err).Msg("error getting next packet")
						break
					} else if err != nil && pkt == nil {
						drain = 0
					}

					if pkt != nil && pkt.StreamIndex() != srcVideoStream.Index() {
						continue
					}

					frames, err = ist.CodecCtx().Decode(pkt)
					if err != nil {
						logger.Error().Err(err).Msg("Fatal error during decoding ")
						break
					}

					// Decode() method doesn't treat EAGAIN and EOF as errors
					// it returns empty frames slice instead. Countinue until
					// input EOF or frames received.
					if len(frames) == 0 && drain < 0 {
						continue
					}

					if frames, err = gmf.DefaultRescaler(swsCtx, frames); err != nil {
						panic(err)
					}

					packets, err := cc.Encode(frames, drain)
					if err != nil {
						logger.Fatal().Err(err).Msg("Error encoding")
					}
					if len(packets) == 0 {
						return
					}

					for _, p := range packets {				
						/*err = ioutil.WriteFile(objectName, p.Data(), 0644)
						if err != nil {
							logger.Fatal().Err(err)
						}*/
						img := new(image.RGBA)
						img.Stride = 4 * cc.Width()
						img.Rect = image.Rect(0, 0, cc.Width(), cc.Height())
						img.Pix = p.Data()
						
						thumbnail = resize.Thumbnail(uint(cfg.Thumbnailer.Width), uint(cfg.Thumbnailer.Height), img, resize.Lanczos3)
														
						p.Free()
						bContinue = false
						continue
					}

					for i, _ := range frames {
						frames[i].Free()
						frameCount++
					}

					if pkt != nil {
						pkt.Free()
						pkt = nil
					}
				}

				for i := 0; i < inputCtx.StreamsCnt(); i++ {
					st, _ := inputCtx.GetStream(i)
					st.CodecCtx().Free()					
					st.Free()
				}

				os.Remove(fileName)
			}
						
			err = jpeg.Encode(&buf, thumbnail, nil /*&jpeg.Options{Quality: 1}*/)
			if err != nil {
				logger.Error().Err(err).Msg("jpeg.Encode")
				return
			}

			b := thumbnail.Bounds()
			width, height := b.Dx(), b.Dy()
				
			contentType := "application/jpeg"
			tags := map[string]string{
				"dimensions": fmt.Sprintf("%dx%d", width, height),
			}
					
			_, err = minioClient.PutObject(ctx, "thumbnail", objectName, bytes.NewReader(buf.Bytes()), int64(len(buf.Bytes())), minio.PutObjectOptions{
				ContentType: contentType,
				UserTags:    tags,
			})
			if err != nil {
				logger.Err(err)
			}		
		}
	}
}

func checkFileExist(bucket, fileName string) bool {
	objInfo, err := minioClient.StatObject(context.Background(), bucket, fileName, minio.StatObjectOptions{})
	if err != nil {
		if err.Error() != "The specified key does not exist." {
			logger.Err(err).Msgf("CheckFileExist bucket : %s, fileName : %s", bucket, fileName)
		}
		return false
	}
	return objInfo.Key != ""
}
