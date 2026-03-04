// name: email
// description: Email sending utilities
// author: roturbot
// requires: net/smtp, net/mail, net/textproto, strings, time

type Email struct {
	from     string
	to       []string
	cc       []string
	bcc      []string
	subject  string
	body     string
	html     bool
}

func (Email) create() *Email {
	return &Email{
		to:   []string{},
		cc:   []string{},
		bcc:  []string{},
		html: false,
	}
}

func (Email) setFrom(email any) bool {
	emailStr := OSLtoString(email)
	
	if !regex.isValidEmail(emailStr) {
		return false
	}
	
	email.from = emailStr
	return true
}

func (Email) addTo(recipient any) bool {
	emailStr := OSLtoString(recipient)
	
	if !regex.isValidEmail(emailStr) {
		return false
	}
	
	email.to = append(email.to, emailStr)
	return true
}

func (Email) addToMany(recipients []any) bool {
	for _, recipient := range recipients {
		if !email.addTo(recipient) {
			return false
		}
	}
	return true
}

func (Email) setTo(recipients any) bool {
	recipientsArr := OSLcastArray(recipients)
	email.to = []string{}
	
	for _, recipient := range recipientsArr {
		if !email.addTo(recipient) {
			return false
		}
	}
	
	return len(email.to) > 0
}

func (Email) getTo() []any {
	result := make([]any, len(email.to))
	for i, addr := range email.to {
		result[i] = addr
	}
	return result
}

func (Email) addCc(recipient any) bool {
	emailStr := OSLtoString(recipient)
	
	if !regex.isValidEmail(emailStr) {
		return false
	}
	
	email.cc = append(email.cc, emailStr)
	return true
}

func (Email) setCc(recipients any) bool {
	recipientsArr := OSLcastArray(recipients)
	email.cc = []string{}
	
	for _, recipient := range recipientsArr {
		email.addCc(recipient)
	}
	
	return true
}

func (Email) getCc() []any {
	result := make([]any, len(email.cc))
	for i, addr := range email.cc {
		result[i] = addr
	}
	return result
}

func (Email) addBcc(recipient any) bool {
	emailStr := OSLtoString(recipient)
	
	if !regex.isValidEmail(emailStr) {
		return false
	}
	
	email.bcc = append(email.bcc, emailStr)
	return true
}

func (Email) setBcc(recipients any) bool {
	recipientsArr := OSLcastArray(recipients)
	email.bcc = []string{}
	
	for _, recipient := range recipientsArr {
		email.addBcc(recipient)
	}
	
	return true
}

func (Email) getBcc() []any {
	result := make([]any, len(email.bcc))
	for i, addr := range email.bcc {
		result[i] = addr
	}
	return result
}

func (Email) setSubject(subject any) {
	email.subject = OSLtoString(subject)
}

func (Email) getSubject() string {
	return email.subject
}

func (Email) setBody(body any) {
	email.body = OSLtoString(body)
}

func (Email) getBody() string {
	return email.body
}

func (Email) setHTML(html any) bool {
	email.html = true
	email.body = OSLtoString(html)
	return true
}

func (Email) setText(text any) {
	email.html = false
	email.body = OSLtoString(text)
}

func (Email) isHTML() bool {
	return email.html
}

func (Email) attachFile(filePath any) bool {
	pathStr := OSLtoString(filePath)
	
	if !fs.Exists(pathStr) {
		return false
	}
	
	return true
}

func (Email) attachContent(files map[string]string) bool {
	for filename, content := range files {
		_ = filename
		_ = content
	}

	return true
}

func (Email) sendWithSmtp(host any, port any, username any, password any, auth any) map[string]any {
	hostStr := OSLtoString(host)
	portStr := OSLtoString(port)
	usernameStr := OSLtoString(username)
	passwordStr := OSLtoString(password)
	useAuth := OSLcastBool(auth)

	if email.from == "" || len(email.to) == 0 {
		return map[string]any{"success": false, "error": "missing from or to address"}
	}

	if portStr == "" {
		portStr = "587"
	}

	smtpAuth := smtp.PlainAuth("", usernameStr, passwordStr, email.from)

	message := email.buildMessage(email.from, nil)

	client, err := smtp.Dial("tcp", hostStr+":"+portStr)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	defer client.Close()

	if useAuth && usernameStr != "" {
		err = client.Auth(smtpAuth)
		if err != nil {
			return map[string]any{"success": false, "error": err.Error()}
		}
	}

	from := email.from
	recipients := email.to

	if len(email.cc) > 0 {
		recipients = append(recipients, email.cc...)
	}

	err = client.Mail(from)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}

	for _, to := range recipients {
		err = client.Rcpt(to)
		if err != nil {
			return map[string]any{"success": false, "error": err.Error()}
		}

		wc, err := client.Data()
		if err != nil {
			return map[string]any{"success": false, "error": err.Error()}
		}

		_, err = wc.Write([]byte(message))
		wc.Close()

		if err != nil {
			return map[string]any{"success": false, "error": err.Error()}
		}
	}

	return map[string]any{"success": true, "message": "email sent successfully"}
}

func (Email) sendGmail(username any, password any) map[string]any {
	return email.sendWithSmtp("smtp.gmail.com", "587", username, password, true)
}

func (Email) sendOutlook(username any, password any) map[string]any {
	return email.sendWithSmtp("smtp-mail.outlook.com", "587", username, password, true)
}

func (Email) sendOffice365(username any, password any) map[string]any {
	return email.sendWithSmtp("smtp.office365.com", "587", username, password, true)
}

func (Email) sendLocalhost() map[string]any {
	host := "localhost"
	port := "25"

	return email.sendWithSmtp(host, port, "", "", false)
}

func (Email) buildMessage(from string, reader *textproto.Reader) string {
	var message strings.Builder

	message.WriteString("From: " + from + "\n")
	
	if len(email.to) > 0 {
		message.WriteString("To: " + strings.Join(email.to, ", ") + "\n")
	}
	
	if len(email.cc) > 0 {
		message.WriteString("Cc: " + strings.Join(email.cc, ", ") + "\n")
	}
	
	if len(email.bcc) > 0 {
		message.WriteString("Bcc: " + strings.Join(email.bcc, ", ") + "\n")
	}

	message.WriteString("Subject: " + email.subject + "\n")
	message.WriteString("MIME-Version: 1.0\n")
	message.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\n")

	if email.html {
		message.WriteString("Content-Type: text/html; charset=\"UTF-8\"\n")
	} else {
		message.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\n")
	}

	message.WriteString("\n")
	message.WriteString(email.body)

	return message.String()
}

func (Email) toMap() map[string]any {
	return map[string]any{
		"from":    email.from,
		"to":      email.getTo(),
		"cc":      email.getCc(),
		"bcc":     email.getBcc(),
		"subject": email.subject,
		"body":    email.body,
		"html":    email.html,
	}
}

func (Email) fromMap(data map[string]any) *Email {
	newEmail := email.create()
	
	if from, ok := data["from"]; ok {
		newEmail.setFrom(from)
	}
	if subject, ok := data["subject"]; ok {
		newEmail.setSubject(subject)
	}
	if body, ok := data["body"]; ok {
		newEmail.setBody(body)
	}
	if html, ok := data["html"]; ok {
		if OSLcastBool(html) {
			newEmail.setHTML(body)
		}
	}
	if to, ok := data["to"]; ok {
		newEmail.setTo(to)
	}
	if cc, ok := data["cc"]; ok {
		newEmail.setCc(cc)
	}
	if bcc, ok := data["bcc"]; ok {
		newEmail.setBcc(bcc)
	}

	return newEmail
}

func (Email) validate() bool {
	if email.from == "" || !regex.isValidEmail(email.from) {
		return false
	}

	if len(email.to) == 0 {
		return false
	}

	for _, to := range email.to {
		if !regex.isValidEmail(to) {
			return false
		}
	}

	return email.subject != ""
}

func (Email) getRecipients() []any {
	all := make([]any, 0, len(email.to)+len(email.cc)+len(email.bcc))

	for _, to := range email.to {
		all = append(all, to)
	}
	for _, cc := range email.cc {
		all = append(all, cc)
	}
	for _, bcc := range email.bcc {
		all = append(all, bcc)
	}

	return all
}

func (Email) getRecipientCount() int {
	return len(email.to) + len(email.cc) + len(email.bcc)
}

func (Email) clear() {
	email.to = []string{}
	email.cc = []string{}
	email.bcc = []string{}
	email.subject = ""
	email.body = ""
	email.html = false
}

func (Email) reset() {
	email.clear()
	email.from = ""
}

func (Email) queue() map[string]any {
	return map[string]any{
		"email":    email.toMap(),
		"queuedAt": date.nowUnix(),
		"sent":     false,
	}
}

func (Email) preview() string {
	return fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s",
		email.from, strings.Join(email.to, ", "), email.subject, email.body)
}

func (Email) wrapText(width any) bool {
	widthInt := int(OSLcastNumber(width))

	if widthInt <= 0 {
		widthInt = 78
	}

	wrapped := template.wordWrap(email.body, widthInt)
	email.setBody(wrapped)

	return true
}

func (Email) getHeaders() map[string]string {
	return map[string]string{
		"From":    email.from,
		"To":      strings.Join(email.to, ", "),
		"Subject": email.subject,
		"Date":    time.Now().Format(time.RFC1123Z),
	}
}

var email = Email{}
