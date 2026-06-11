package tray

var wizardHTML = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>校园网自动登录器 - 配置向导</title>
<style>
body { font-family: Microsoft YaHei, sans-serif; background: #1e1e2e; color: #cdd6f4; padding: 20px; max-width: 700px; margin: 0 auto; }
h2 { color: #89b4fa; border-bottom: 2px solid #313244; padding-bottom: 10px; }
.step { background: #313244; padding: 15px; border-radius: 8px; margin: 15px 0; }
.step-title { color: #89b4fa; font-weight: bold; margin-bottom: 10px; font-size: 16px; }
.btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; margin: 5px; }
.btn-primary { background: #89b4fa; color: #1e1e2e; }
.btn-success { background: #a6e3a1; color: #1e1e2e; }
.btn-warning { background: #f9e2af; color: #1e1e2e; }
.btn:hover { opacity: 0.8; }
.result { padding: 10px; margin: 10px 0; border-radius: 4px; }
.result-ok { background: #a6e3a122; color: #a6e3a1; }
.result-err { background: #f38ba822; color: #f38ba8; }
.result-warn { background: #f9e2af22; color: #f9e2af; }
input[type=text], input[type=password], textarea { width: 100%; padding: 8px; background: #45475a; color: #cdd6f4; border: 1px solid #585b70; border-radius: 4px; box-sizing: border-box; }
textarea { height: 80px; font-family: Consolas, monospace; font-size: 12px; }
label { display: block; margin: 10px 0 5px; color: #a6adc8; }
.spinner { display: inline-block; width: 16px; height: 16px; border: 2px solid #585b70; border-top: 2px solid #89b4fa; border-radius: 50%; animation: spin 1s linear infinite; }
@keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
.tutorial { background: #45475a; padding: 12px; border-radius: 6px; margin: 10px 0; font-size: 13px; line-height: 1.6; }
.tutorial code { background: #585b70; padding: 2px 6px; border-radius: 3px; color: #f9e2af; }
.hidden { display: none; }
.step-num { display: inline-block; width: 24px; height: 24px; background: #89b4fa; color: #1e1e2e; border-radius: 50%; text-align: center; line-height: 24px; font-weight: bold; margin-right: 8px; }
</style></head><body>
<h2>🔧 校园网自动登录器 - 配置向导</h2>

<!-- 步骤 1: 自动检测 -->
<div class="step" id="step1">
<div class="step-title"><span class="step-num">1</span>自动检测</div>
<p>程序将自动检测校园网认证系统。请确保已连接校园网。</p>
<button class="btn btn-primary" onclick="startDetect()">开始检测</button>
<div id="detect-result"></div>
</div>

<!-- 步骤 2: 输入账号密码 -->
<div class="step" id="step2">
<div class="step-title"><span class="step-num">2</span>输入账号密码</div>
<label>认证服务器</label>
<input type="text" id="server" value="" placeholder="例如: 210.44.114.32:801">
<label>账号</label>
<input type="text" id="username" value="" placeholder="请输入校园网账号">
<label>密码</label>
<input type="password" id="password" value="" placeholder="请输入密码">
<label>运营商</label>
<select id="carrier" style="width:100%;padding:8px;background:#45475a;color:#cdd6f4;border:1px solid #585b70;border-radius:4px;">
<option value="campus">校园用户</option>
<option value="dx">校园电信 (@dx)</option>
<option value="lt">校园联通 (@lt)</option>
<option value="other">校园其他</option>
</select>
<div style="margin-top:15px;">
<button class="btn btn-success" onclick="testLogin()">测试登录</button>
</div>
<div id="test-result"></div>
</div>

<!-- 步骤 3: 登录失败 - 抓包引导 -->
<div class="step" id="step3">
<div class="step-title"><span class="step-num">3</span>登录失败？获取正确的登录 API</div>

<div class="result result-warn" id="fail-reason"></div>

<p>默认的登录 API 不适用于你的学校，请按以下步骤手动抓取：</p>

<div class="tutorial">
<p><strong>📋 抓包步骤：</strong></p>
<p>1. 打开浏览器，按 <code>F12</code> 打开开发者工具</p>
<p>2. 点击 <code>Network（网络）</code> 标签</p>
<p>3. <span style="color:#f9e2af">⚠️ 勾选「Preserve log（保留日志）」</span>（重要！）</p>
<p>4. 在地址栏访问任意网站（如 <code>baidu.com</code>），会被重定向到登录页面</p>
<p>5. 输入账号密码，点击登录</p>
<p>6. 在 Network 列表中找到包含 <code>login</code> 的请求</p>
<p>7. 右键该请求 → <code>Copy（复制）</code> → <code>Copy as cURL（以 cURL 格式复制）</code></p>
<p>8. 选择 <code>Copy as cURL (bash)</code> 或 <code>Copy as cURL (cmd)</code> 都可以</p>
</div>

<label>粘贴 cURL 命令：</label>
<textarea id="curl-input" placeholder="粘贴从浏览器复制的 cURL 命令..."></textarea>

<div style="margin-top:10px;">
<button class="btn btn-warning" onclick="parseCurl()">解析 cURL</button>
</div>
<div id="parse-result"></div>
</div>

<!-- 步骤 4: 使用自定义 API 测试 -->
<div class="step" id="step4">
<div class="step-title"><span class="step-num">4</span>测试自定义登录 API</div>

<label>解析出的登录 URL：</label>
<input type="text" id="custom-api" value="" placeholder="登录 API URL">

<div style="margin-top:10px;">
<button class="btn btn-success" onclick="testCustomLogin()">测试登录</button>
<button class="btn btn-primary" onclick="saveCustomConfig()">保存配置</button>
</div>
<div id="custom-test-result"></div>
</div>

<!-- 步骤 5: 配置完成 -->
<div class="step" id="step5">
<div class="step-title">✅ 配置完成</div>
<p>配置已保存！程序将自动保活校园网连接。</p>
<p>您可以关闭此页面，程序会在后台运行。</p>
</div>

<script>
let detectData = {};

async function startDetect() {
    document.getElementById('detect-result').innerHTML = '<div class="result result-warn"><span class="spinner"></span> 正在检测...</div>';

    try {
        const resp = await fetch('/api/detect');
        const data = await resp.json();
        detectData = data;

        if (data.found) {
            document.getElementById('detect-result').innerHTML =
                '<div class="result result-ok">✓ 检测成功！<br>认证系统: ' + data.system + '<br>服务器: ' + data.server + ':' + data.port + '</div>';
            document.getElementById('server').value = data.server + ':' + data.port;
            // 步骤已默认显示
        } else {
            document.getElementById('detect-result').innerHTML =
                '<div class="result result-err">✗ 未检测到认证系统。请确保已连接校园网，或手动输入服务器地址。</div>';
            // 步骤已默认显示
        }
    } catch (e) {
        document.getElementById('detect-result').innerHTML =
            '<div class="result result-err">✗ 检测失败: ' + e.message + '</div>';
        document.getElementById('step2').classList.remove('hidden');
    }
}

async function testLogin() {
    const server = document.getElementById('server').value;
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const carrier = document.getElementById('carrier').value;

    if (!username || !password) {
        document.getElementById('test-result').innerHTML = '<div class="result result-err">请输入账号和密码</div>';
        return;
    }

    document.getElementById('test-result').innerHTML = '<div class="result result-warn"><span class="spinner"></span> 正在测试登录...</div>';

    try {
        const resp = await fetch('/api/test-login', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({server, username, password, carrier})
        });
        const data = await resp.json();

        if (data.success) {
            document.getElementById('test-result').innerHTML = '<div class="result result-ok">✓ 登录成功！<br>' + data.message + '</div>';
            // 保存配置
            saveConfig();
        } else {
            document.getElementById('test-result').innerHTML = '<div class="result result-err">✗ 登录失败</div>';
            // 显示抓包引导
            document.getElementById('fail-reason').innerHTML = '默认 API 不适用于你的学校，错误信息: ' + data.message;
            // 步骤已默认显示
        }
    } catch (e) {
        document.getElementById('test-result').innerHTML = '<div class="result result-err">✗ 请求失败: ' + e.message + '</div>';
    }
}

async function parseCurl() {
    const curlCmd = document.getElementById('curl-input').value;

    if (!curlCmd || !curlCmd.includes('curl')) {
        document.getElementById('parse-result').innerHTML = '<div class="result result-err">请粘贴完整的 curl 命令</div>';
        return;
    }

    document.getElementById('parse-result').innerHTML = '<div class="result result-warn"><span class="spinner"></span> 正在解析...</div>';

    try {
        const resp = await fetch('/api/parse-curl', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({curl: curlCmd})
        });
        const data = await resp.json();

        if (data.success) {
            document.getElementById('parse-result').innerHTML =
                '<div class="result result-ok">✓ 解析成功！<br>URL: ' + data.url + '<br>方法: ' + data.method + '</div>';
            document.getElementById('custom-api').value = data.url;
            // 步骤已默认显示
        } else {
            document.getElementById('parse-result').innerHTML =
                '<div class="result result-err">✗ ' + data.message + '</div>';
        }
    } catch (e) {
        document.getElementById('parse-result').innerHTML =
            '<div class="result result-err">✗ 请求失败: ' + e.message + '</div>';
    }
}

async function testCustomLogin() {
    const apiURL = document.getElementById('custom-api').value;
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    if (!apiURL) {
        document.getElementById('custom-test-result').innerHTML = '<div class="result result-err">请填写登录 API URL</div>';
        return;
    }

    document.getElementById('custom-test-result').innerHTML = '<div class="result result-warn"><span class="spinner"></span> 正在测试...</div>';

    try {
        const resp = await fetch('/api/test-custom-login', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({url: apiURL, username, password})
        });
        const data = await resp.json();

        if (data.success) {
            document.getElementById('custom-test-result').innerHTML =
                '<div class="result result-ok">✓ 登录成功！<br>' + data.message + '</div>';
        } else {
            document.getElementById('custom-test-result').innerHTML =
                '<div class="result result-err">✗ 登录失败: ' + data.message + '</div>';
        }
    } catch (e) {
        document.getElementById('custom-test-result').innerHTML =
            '<div class="result result-err">✗ 请求失败: ' + e.message + '</div>';
    }
}

async function saveConfig() {
    const server = document.getElementById('server').value;
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const carrier = document.getElementById('carrier').value;
    const customAPI = document.getElementById('custom-api') ? document.getElementById('custom-api').value : '';

    if (!username || !password) {
        document.getElementById('test-result').innerHTML = '<div class="result result-err">请输入账号和密码</div>';
        return;
    }

    // 显示正在保存
    document.getElementById('test-result').innerHTML = '<div class="result result-warn"><span class="spinner"></span> 正在保存...</div>';

    try {
        const resp = await fetch('/api/save-config', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({server, username, password, carrier, custom_api: customAPI})
        });
        const data = await resp.json();

        if (data.success) {
            document.getElementById('test-result').innerHTML = '<div class="result result-ok">✓ 配置已保存！程序将使用新配置。</div>';
            // 跳转到完成步骤
            setTimeout(() => {
                document.getElementById('step5').scrollIntoView({behavior: 'smooth'});
            }, 1000);
        } else {
            document.getElementById('test-result').innerHTML = '<div class="result result-err">✗ 保存失败: ' + data.message + '</div>';
        }
    } catch (e) {
        document.getElementById('test-result').innerHTML = '<div class="result result-err">✗ 请求失败: ' + e.message + '</div>';
    }
}

async function saveCustomConfig() {
    const server = document.getElementById('server').value;
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const carrier = document.getElementById('carrier').value;
    const customAPI = document.getElementById('custom-api').value;

    if (!customAPI) {
        document.getElementById('custom-test-result').innerHTML = '<div class="result result-err">请先解析 cURL 命令</div>';
        return;
    }

    document.getElementById('custom-test-result').innerHTML = '<div class="result result-warn"><span class="spinner"></span> 正在保存...</div>';

    try {
        const resp = await fetch('/api/save-config', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({server, username, password, carrier, custom_api: customAPI})
        });
        const data = await resp.json();

        if (data.success) {
            // 配置已保存
        } else {
            document.getElementById('custom-test-result').innerHTML =
                '<div class="result result-err">✗ 保存失败: ' + data.message + '</div>';
        }
    } catch (e) {
        document.getElementById('custom-test-result').innerHTML =
            '<div class="result result-err">✗ 请求失败: ' + e.message + '</div>';
    }
}
</script>
</body></html>`
