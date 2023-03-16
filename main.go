package main

import (
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func main() {
	// old way
	config := aws.NewConfig().WithRegion("eu-central-1")
	sess, err := session.NewSession(config)
	if err != nil {
		log.Fatalf("main session %s", err)
	}

	svc := sts.New(sess)
	result, err := svc.AssumeRole(&sts.AssumeRoleInput{
		RoleArn:         aws.String(os.Getenv("KOPS_ROLE_ARN")),
		RoleSessionName: aws.String("kopsrole"),
		DurationSeconds: aws.Int64(60 * 60 * 1), // 1 hours, due to role chaining its aws maximum
	})
	if err != nil {
		log.Fatalf("assume role %s", err)
	}

	go aliveCheck(aws.StringValue(result.Credentials.AccessKeyId), aws.StringValue(result.Credentials.SecretAccessKey), aws.StringValue(result.Credentials.SessionToken))

	aliveCheckNew()
}

// kops call with assumed role
func aliveCheck(accessid, secretkey, token string) {
	config := &aws.Config{
		Region:      aws.String("eu-central-1"),
		Credentials: credentials.NewStaticCredentials(accessid, secretkey, token),
	}

	sess, err := session.NewSession(config)
	if err != nil {
		log.Fatalf("aliveCheck session %s", err)
	}

	requestLogger := newRequestLogger()

	stsClient := sts.New(sess, config)
	stsClient.Handlers.Send.PushFront(requestLogger)

	// we can think that kops cli is executed in loop "enough long"
	for {
		res, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		if err != nil {
			log.Fatalf("aliveCheck loop %s", err)
		}
		log.Printf("old %+v", res)
		time.Sleep(1 * time.Minute)
	}
}

func aliveCheckNew() {
	config := aws.NewConfig().WithRegion("eu-central-1")

	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            *config,
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		log.Fatalf("aliveCheckNew sess %s", err)
	}

	roleARN := os.Getenv("KOPS_ROLE_ARN")
	if roleARN != "" {
		creds := stscreds.NewCredentials(sess, roleARN)
		config = &aws.Config{Credentials: creds}
		config = config.WithRegion("eu-central-1")
	}
	requestLogger := newRequestLogger()

	stsClient := sts.New(sess, config)
	stsClient.Handlers.Send.PushFront(requestLogger)

	// we can think that kops cli is executed in loop "enough long"
	for {
		res, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		if err != nil {
			log.Fatalf("aliveCheckNew loop %s", err)
		}
		log.Printf("new %+v", res)
		time.Sleep(1 * time.Minute)
	}
}

// RequestLogger logs every AWS request
type RequestLogger struct {
}

func newRequestLogger() func(r *request.Request) {
	rl := &RequestLogger{}
	return rl.log
}

// Handler for aws-sdk-go that logs all requests
func (l *RequestLogger) log(r *request.Request) {
	service := r.ClientInfo.ServiceName
	name := "?"
	if r.Operation != nil {
		name = r.Operation.Name
	}
	methodDescription := service + "/" + name
	log.Printf("AWS request: %s", methodDescription)
}
