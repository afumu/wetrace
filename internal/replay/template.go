package replay

const chatTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body {
  width: {{.Width}}px;
  height: {{.Height}}px;
  background: {{.Background}};
  font-family: -apple-system, "Microsoft YaHei", sans-serif;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  padding: 20px;
}
.msg { display: flex; margin-bottom: 12px; align-items: flex-start; }
.msg.self { flex-direction: row-reverse; }
.avatar {
  width: 40px; height: 40px; border-radius: 6px;
  background: #7c8dff; color: #fff; display: flex;
  align-items: center; justify-content: center;
  font-size: 16px; flex-shrink: 0;
}
.msg.self .avatar { background: #57c457; }
.bubble-wrap { max-width: 65%; margin: 0 10px; }
.name { font-size: 12px; color: #888; margin-bottom: 3px; }
.msg.self .name { text-align: right; }
.bubble {
  background: #fff; border-radius: 8px; padding: 10px 14px;
  font-size: 15px; line-height: 1.5; word-break: break-all;
  box-shadow: 0 1px 2px rgba(0,0,0,0.08);
}
.msg.self .bubble { background: #95ec69; }
.time { font-size: 11px; color: #aaa; margin-top: 2px; }
.msg.self .time { text-align: right; }
</style>
</head>
<body>
{{range .Messages}}
<div class="msg{{if .IsSelf}} self{{end}}">
  <div class="avatar">{{.AvatarText}}</div>
  <div class="bubble-wrap">
    <div class="name">{{.SenderName}}</div>
    <div class="bubble">{{.Content}}</div>
    <div class="time">{{.Time}}</div>
  </div>
</div>
{{end}}
</body>
</html>`
