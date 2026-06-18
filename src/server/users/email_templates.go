package users

const (
	InvitationEmailSubject    = "User invitation"
	PasswordResetEmailSubject = "Password reset"

	DefaultInvitationEmailTemplate = `Hello,

You've been invited to join {{.ServerUrl}}. An account has been created for you with the following username:

Username: {{.Username}}

To get started, please set your password using the link below:

{{.InvitationLink}}

This invitation link is valid until {{.ExpirationDate}}.

If you were not expecting this invitation, you can safely ignore this email.

Best regards
` // #nosec G101 (CWE-798): Potential hardcoded credentials

	DefaultPasswordResetEmailBody = `Hello,

A password reset was requested for your account on %s.

Username: %s

To set a new password, please use the link below:

%s

This password reset link is valid until %s.

If you were not expecting this password reset, you can safely ignore this email.

Best regards
` // #nosec G101 (CWE-798): Potential hardcoded credentials
)
