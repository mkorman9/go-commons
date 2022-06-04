package mail

type Props = map[string]interface{}

type Backend interface {
	SendMail(targetName, targetEmail, template string, props Props) error
}
