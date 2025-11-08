package mail

import "fmt"

func EmailVerificationTemplate(username, code string) string {
	return fmt.Sprintf(`
  <!DOCTYPE html>
  <html>
  <head>
      <style>
          body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
          .container { max-width: 600px; margin: 0 auto; padding: 20px; }
          .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
          .content { padding: 20px; background-color: #f9f9f9; }
          .code { font-size: 32px; font-weight: bold; color: #4CAF50; text-align: center; padding: 20px; background-color: white; margin: 20px 0; border-radius: 
  5px; }
          .footer { text-align: center; padding: 20px; font-size: 12px; color: #888; }
      </style>
  </head>
  <body>
      <div class="container">
          <div class="header">
              <h1>GoMall - 邮箱验证</h1>
          </div>
          <div class="content">
              <p>你好 %s,</p>
              <p>感谢注册 GoMall！请使用以下验证码完成邮箱验证：</p>
              <div class="code">%s</div>
              <p>此验证码将在 <strong>15 分钟</strong>后过期。</p>
              <p>如果你没有注册 GoMall 账户，请忽略此邮件。</p>
          </div>
          <div class="footer">
              <p>&copy; 2024 GoMall. All rights reserved.</p>
          </div>
      </div>
  </body>
  </html>
  `, username, code)
}

// PasswordResetTemplate 密码重置邮件模板
func PasswordResetTemplate(username, code string) string {
	return fmt.Sprintf(`
  <!DOCTYPE html>
  <html>
  <head>
      <style>
          body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
          .container { max-width: 600px; margin: 0 auto; padding: 20px; }
          .header { background-color: #FF9800; color: white; padding: 20px; text-align: center; }
          .content { padding: 20px; background-color: #f9f9f9; }
          .code { font-size: 32px; font-weight: bold; color: #FF9800; text-align: center; padding: 20px; background-color: white; margin: 20px 0; border-radius: 
  5px; }
          .warning { background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 10px; margin: 20px 0; }
          .footer { text-align: center; padding: 20px; font-size: 12px; color: #888; }
      </style>
  </head>
  <body>
      <div class="container">
          <div class="header">
              <h1>GoMall - 密码重置</h1>
          </div>
          <div class="content">
              <p>你好 %s,</p>
              <p>我们收到了你的密码重置请求。请使用以下验证码重置密码：</p>
              <div class="code">%s</div>
              <p>此验证码将在 <strong>15 分钟</strong>后过期。</p>
              <div class="warning">
                  <strong>⚠️ 安全提示：</strong>
                  <p>如果你没有请求重置密码，请立即忽略此邮件并确保你的账户安全。</p>
              </div>
          </div>
          <div class="footer">
              <p>&copy; 2024 GoMall. All rights reserved.</p>
          </div>
      </div>
  </body>
  </html>
  `, username, code)
}
