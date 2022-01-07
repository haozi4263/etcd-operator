package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/haozi4263/etcd-operator/pkg/file"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/snapshot"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"time"
)

func logErr(log logr.Logger, err error, message string) error {
	log.Error(err, message)
	return fmt.Errorf("%s: %s", message, err)
}

func main() {
	var (
		etcdURL            string
		backupTempDir      string
		timeoutSeconds     int64 //备份超时时间
		dialTimeoutSeconds int64 // 访问超时时间
	)

	flag.StringVar(&backupTempDir, "backup-tmp-dir", os.TempDir(), "The Directory to temp place backup etcd cluster")
	flag.StringVar(&etcdURL, "etcd-url", "", "URL for backup etcd.")
	flag.Int64Var(&dialTimeoutSeconds, "dial-timeout-seconds", 5, "Timeout for fialing the Etcd")
	flag.Int64Var(&timeoutSeconds, "timeout-seconds", 60, "Timeout for Backup the Etcd")
	flag.Parse()
	zapLogger := zap.NewRaw(zap.UseDevMode(true))
	ctrl.SetLogger(zapr.NewLogger(zapLogger))

	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutSeconds))
	defer ctxCancel()

	log := ctrl.Log.WithName("backup")
	log.Info("Connecttiong to Etcd and getting Snapshot date")

	// 定义一个本地的数据目录
	localPath := filepath.Join(backupTempDir, "snapshot.db")
	// 创建etcd snapshot manager
	etcdManager := snapshot.NewV3(zapLogger)
	//保存etcd snapshot数据到localPath
	err := etcdManager.Save(ctx, clientv3.Config{
		Endpoints:   []string{etcdURL},
		DialTimeout: time.Second * time.Duration(dialTimeoutSeconds),
	}, localPath)
	if err != nil {
		panic(logErr(log, err, "failed to get etcd snapshot date"))
	}

	// 数据保存本地成功，上传到对象存储minio/oss中
	// todo ,根据传递进来的参数判断初始化s3还是oss
	endpoint := "play.min.io"
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	s3Uploader := file.News3Uploader(endpoint, accessKeyID, secretAccessKey)

	log.Info("Uploading snapshot...")
	size, err := s3Uploader.Uploader(ctx, localPath)
	if err != nil {
		panic(logErr(log, err, "failed to upload backup etcd"))
	}
	log.WithValues("uploader-size", size).Info("Backup completed")
}
