(function(){let e=document.createElement(`link`).relList;if(e&&e.supports&&e.supports(`modulepreload`))return;for(let e of document.querySelectorAll(`link[rel="modulepreload"]`))n(e);new MutationObserver(e=>{for(let t of e)if(t.type===`childList`)for(let e of t.addedNodes)e.tagName===`LINK`&&e.rel===`modulepreload`&&n(e)}).observe(document,{childList:!0,subtree:!0});function t(e){let t={};return e.integrity&&(t.integrity=e.integrity),e.referrerPolicy&&(t.referrerPolicy=e.referrerPolicy),e.crossOrigin===`use-credentials`?t.credentials=`include`:e.crossOrigin===`anonymous`?t.credentials=`omit`:t.credentials=`same-origin`,t}function n(e){if(e.ep)return;e.ep=!0;let n=t(e);fetch(e.href,n)}})();var e=`modulepreload`,t=function(e){return`/`+e},n={},r=function(r,i,a){let o=Promise.resolve();if(i&&i.length>0){let r=document.getElementsByTagName(`link`),s=document.querySelector(`meta[property=csp-nonce]`),c=s?.nonce||s?.getAttribute(`nonce`);function l(e){return Promise.all(e.map(e=>Promise.resolve(e).then(e=>({status:`fulfilled`,value:e}),e=>({status:`rejected`,reason:e}))))}o=l(i.map(i=>{if(i=t(i,a),i in n)return;n[i]=!0;let o=i.endsWith(`.css`),s=o?`[rel="stylesheet"]`:``;if(a)for(let e=r.length-1;e>=0;e--){let t=r[e];if(t.href===i&&(!o||t.rel===`stylesheet`))return}else if(document.querySelector(`link[href="${i}"]${s}`))return;let l=document.createElement(`link`);if(l.rel=o?`stylesheet`:e,o||(l.as=`script`),l.crossOrigin=``,l.href=i,c&&l.setAttribute(`nonce`,c),document.head.appendChild(l),o)return new Promise((e,t)=>{l.addEventListener(`load`,e),l.addEventListener(`error`,()=>t(Error(`Unable to preload CSS for ${i}`)))})}))}function s(e){let t=new Event(`vite:preloadError`,{cancelable:!0});if(t.payload=e,window.dispatchEvent(t),!t.defaultPrevented)throw e}return o.then(e=>{for(let t of e||[])t.status===`rejected`&&s(t.reason);return r().catch(s)})},i=`0.1.0`,a=document.querySelector(`#app`);if(!a)throw Error(`missing #app`);var o=a,s={running:!1,stateLabel:`未运行`,address:``,config:{rootPath:`/Users/me/Downloads`,port:9090,maxLevel:0,isAllowUpload:!1,isSecure:!1,scratchMaxItems:500,scratchMaxItemSize:`10MB`},hasCredentials:!1,configPath:`~/Library/Application Support/tiny/config.yml`,accessLogPath:`~/Library/Application Support/tiny/access.log`,portRestartRequired:!1,lastError:``},c=[{value:``,label:`全部`},{value:`access`,label:`access`},{value:`download`,label:`download`},{value:`upload`,label:`upload`},{value:`login`,label:`login`},{value:`reject`,label:`reject`},{value:`error`,label:`error`}],l=s,u=[],d=ie(),f={},p=`panel`,m=``,h=!1,g=null,_=v();S();function v(){let e=()=>r(()=>import(`./service-BjP_ED7t.js`),[]),t=async(t,n,r)=>{let i=await y(e);return i?(h=!1,await r(i)):(h=!0,b(t,n))};return{GetStatus:()=>t(`GetStatus`,[],e=>e.GetStatus()),StartSharing:()=>t(`StartSharing`,[],e=>e.StartSharing()),StopSharing:()=>t(`StopSharing`,[],e=>e.StopSharing()),UpdateConfig:e=>t(`UpdateConfig`,[e],t=>t.UpdateConfig(e)),SetCredentials:e=>t(`SetCredentials`,[e],t=>t.SetCredentials(e)),GetLogs:e=>t(`GetLogs`,[e],t=>t.GetLogs(e)),ClearLogs:()=>t(`ClearLogs`,[],e=>e.ClearLogs()),ChooseDirectory:e=>t(`ChooseDirectory`,[e],t=>t.ChooseDirectory(e)),ExportLogs:e=>t(`ExportLogs`,[e],t=>t.ExportLogs(e)),OpenConfigDir:()=>t(`OpenConfigDir`,[],e=>e.OpenConfigDir())}}async function y(e){try{return await e()}catch{return null}}async function b(e,t){switch(await new Promise(e=>window.setTimeout(e,120)),e){case`GetStatus`:return l;case`StartSharing`:return l={...l,running:!0,stateLabel:`运行中`,address:q(l.config.port),lastError:``},l;case`StopSharing`:return l={...l,running:!1,stateLabel:`未运行`,address:``,lastError:``},l;case`UpdateConfig`:{let e=t[0];if(l.running&&e.port!==void 0&&e.port!==l.config.port&&!e.restartPort)throw l={...l,lastError:`修改端口需要确认并重启服务`},Error(l.lastError);let n=x(l.config,e);return l={...l,config:n,address:l.running?q(n.port):``,portRestartRequired:!1,lastError:``},l}case`ChooseDirectory`:return`/Users/me/Shared`;case`GetLogs`:return ae(d,t[0]??{});case`ClearLogs`:d=[];return;case`ExportLogs`:return`/Users/me/Downloads/onetiny-access.csv`;case`OpenConfigDir`:return;case`SetCredentials`:{let e=t[0];if(!e.username.trim())throw Error(`用户名不能为空`);if(!e.password.trim())throw Error(`密码不能为空`);if(e.password!==e.confirmPassword)throw Error(`两次输入的密码不一致`);return l={...l,config:{...l.config,isSecure:e.enableSecure?!0:l.config.isSecure},hasCredentials:!0,lastError:``},l}default:throw Error(`unknown mock method: ${e}`)}}function x(e,t){let n={...e};return t.rootPath!=null&&(n.rootPath=t.rootPath),t.port!=null&&(n.port=t.port),t.maxLevel!=null&&(n.maxLevel=t.maxLevel),t.isAllowUpload!=null&&(n.isAllowUpload=t.isAllowUpload),t.isSecure!=null&&(n.isSecure=t.isSecure),t.scratchMaxItems!=null&&(n.scratchMaxItems=t.scratchMaxItems),t.scratchMaxItemSize!=null&&(n.scratchMaxItemSize=t.scratchMaxItemSize),n}async function S(){try{l=await _.GetStatus(),p===`logs`&&(u=await _.GetLogs(f)),m=W()}catch(e){m=$(e)}C()}function C(){let e=m||l.lastError;o.innerHTML=`
    <main class="shell">
      <section class="top-control" aria-label="共享状态">
        <div class="access-block">
          <div class="access-labels">
            <span class="label">访问地址</span>
            <span class="state ${l.running?`state-running`:``}">${Q(l.stateLabel)}</span>
          </div>
          <code>${Q(l.address||`服务未启动`)}</code>
        </div>
        <div class="top-actions">
          <button data-action="copy" ${l.address?``:`disabled`}>复制地址</button>
          <button class="primary" data-action="${l.running?`stop`:`start`}">
            ${l.running?`停止共享`:`启动共享`}
          </button>
        </div>
      </section>

      ${e?`<p class="notice">${Q(e)}</p>`:``}

      <header class="app-header">
        <div class="brand">
          <div class="mark">O</div>
          <div>
            <h1>OneTiny</h1>
            <p>局域网文件共享控制面板</p>
          </div>
        </div>
      </header>

      <nav class="tabs">
        ${w(`panel`,`控制面板`)}
        ${w(`security`,`安全设置`)}
        ${w(`logs`,`访问日志`)}
        ${w(`about`,`关于`)}
      </nav>

      <section class="content">
        ${ee()}
      </section>
      ${k()}
    </main>
  `,j()}function w(e,t){return`<button class="tab ${p===e?`active`:``}" data-tab="${e}">${t}</button>`}function ee(){switch(p){case`panel`:return te();case`security`:return T();case`logs`:return D();case`about`:return O()}}function te(){return`
    <div class="control-list">
      <label class="control-row directory-row">
        <span>共享目录</span>
        <input class="readonly-input" type="text" value="${Q(l.config.rootPath)}" readonly>
        <button type="button" data-action="choose-dir">选择</button>
      </label>

      <div class="control-row">
        <span>允许上传</span>
        <label class="switch">
          <input type="checkbox" data-toggle="upload" ${l.config.isAllowUpload?`checked`:``}>
          <span></span>
        </label>
      </div>

      ${E()}

      <label class="control-row">
        <span>端口</span>
        <input class="number-input" type="number" min="1" max="65535" step="1" value="${l.config.port}" data-number="port">
      </label>

      <label class="control-row">
        <span>最大访问层级</span>
        <input class="number-input" type="number" min="0" max="255" step="1" value="${l.config.maxLevel}" data-number="maxLevel">
      </label>

      <label class="control-row">
        <span>临时列表容量</span>
        <input class="number-input" type="number" min="1" max="10000" step="1" value="${l.config.scratchMaxItems}" data-number="scratchMaxItems">
      </label>

      <label class="control-row">
        <span>单条大小上限</span>
        <input class="number-input" type="text" value="${Q(l.config.scratchMaxItemSize)}" data-text-setting="scratchMaxItemSize">
      </label>
    </div>
  `}function T(){return`
    <div class="control-list">
      ${E()}
      <div class="control-row">
        <span>账号状态</span>
        <strong class="value-pill ${l.hasCredentials?`ok`:``}">
          ${l.hasCredentials?`已配置`:`未配置`}
        </strong>
      </div>
      <div class="control-row">
        <span>登录保护</span>
        <strong class="value-pill ${l.config.isSecure?`ok`:``}">
          ${l.config.isSecure?`已开启`:`已关闭`}
        </strong>
      </div>
    </div>
  `}function E(){return`
    <div class="control-row">
      <span>登录保护</span>
      <div class="inline-actions">
        <label class="switch">
          <input type="checkbox" data-toggle="secure" ${l.config.isSecure?`checked`:``}>
          <span></span>
        </label>
        <button type="button" data-action="credentials">账号设置</button>
      </div>
    </div>
  `}function D(){return`
    <form class="log-filters" aria-label="访问日志筛选">
      <label>
        <span>事件</span>
        <select name="event">
          ${c.map(e=>`
                <option value="${Q(e.value)}" ${e.value===(f.event??``)?`selected`:``}>
                  ${Q(e.label)}
                </option>
              `).join(``)}
        </select>
      </label>
      <label>
        <span>开始时间</span>
        <input name="since" type="datetime-local" value="${Q(X(f.since))}">
      </label>
      <label>
        <span>结束时间</span>
        <input name="until" type="datetime-local" value="${Q(X(f.until))}">
      </label>
      <div class="toolbar">
        <button type="button" data-action="refresh-logs">刷新</button>
        <button type="button" data-action="export-logs">导出 CSV</button>
        <button type="button" class="danger" data-action="clear-logs">清空</button>
      </div>
    </form>
    <div class="log-table">
      ${A()}
    </div>
  `}function O(){return`
    <div class="about-panel">
      <dl class="about">
        <dt>版本</dt>
        <dd>OneTiny GUI ${Q(i)}</dd>
        <dt>模式</dt>
        <dd>${Q(K(G()))}</dd>
        <dt>配置文件</dt>
        <dd>${Q(l.configPath||`-`)}</dd>
        <dt>访问日志</dt>
        <dd>${Q(l.accessLogPath||`-`)}</dd>
      </dl>
      <button data-action="open-config">打开配置目录</button>
    </div>
  `}function k(){return g?`
    <dialog class="credential-dialog" aria-labelledby="credential-title">
      <form class="credential-form" method="dialog">
        <div class="dialog-header">
          <h2 id="credential-title">账号设置</h2>
          <button class="icon-button" type="button" data-action="close-credentials" aria-label="关闭">×</button>
        </div>
        ${g.error?`<p class="dialog-error">${Q(g.error)}</p>`:``}
        <label>
          <span>用户名</span>
          <input name="username" autocomplete="username" value="${Q(g.username)}">
        </label>
        <label>
          <span>密码</span>
          <input name="password" type="password" autocomplete="new-password" value="${Q(g.password)}">
        </label>
        <label>
          <span>确认密码</span>
          <input name="confirmPassword" type="password" autocomplete="new-password" value="${Q(g.confirmPassword)}">
        </label>
        <div class="dialog-actions">
          <button type="button" data-action="close-credentials">取消</button>
          <button class="primary" type="submit">保存</button>
        </div>
      </form>
    </dialog>
  `:``}function A(){return u.length===0?`<p class="empty">暂无访问日志</p>`:`
    <table>
      <thead>
        <tr>
          <th class="log-time">时间</th>
          <th class="log-ip">客户端 IP</th>
          <th class="log-method">方法</th>
          <th class="log-event">事件</th>
          <th class="log-path">路径</th>
          <th class="log-status">状态</th>
          <th class="log-result">结果</th>
        </tr>
      </thead>
      <tbody>
        ${u.map(e=>`
              <tr>
                <td class="log-time">${Q(re(e.time))}</td>
                <td>${Q(e.clientIP)}</td>
                <td>${Q(e.method||`-`)}</td>
                <td>${Q(e.event)}</td>
                <td class="log-path">${Q(e.path||`-`)}</td>
                <td>${Q(e.status?String(e.status):`-`)}</td>
                <td>${Q(e.result||`-`)}</td>
              </tr>
            `).join(``)}
      </tbody>
    </table>
  `}function j(){o.querySelectorAll(`[data-tab]`).forEach(e=>{e.addEventListener(`click`,()=>{p=e.dataset.tab,S()})}),o.querySelector(`[data-action="start"]`)?.addEventListener(`click`,()=>{M(async()=>{l=await _.StartSharing(),m=W(),C()})}),o.querySelector(`[data-action="stop"]`)?.addEventListener(`click`,()=>{M(async()=>{l=await _.StopSharing(),m=W(),C()})}),o.querySelector(`[data-action="copy"]`)?.addEventListener(`click`,()=>{M(async()=>{l.address&&(await navigator.clipboard.writeText(l.address),m=`访问地址已复制`,C())})}),o.querySelector(`[data-action="choose-dir"]`)?.addEventListener(`click`,()=>{M(async()=>{let e=await _.ChooseDirectory(l.config.rootPath);e&&(l=await _.UpdateConfig({rootPath:e}),m=W(),C())})}),o.querySelectorAll(`[data-toggle="upload"]`).forEach(e=>{e.addEventListener(`change`,()=>{M(async()=>{l=await _.UpdateConfig({isAllowUpload:e.checked}),m=W(),C()})})}),o.querySelectorAll(`[data-toggle="secure"]`).forEach(e=>{e.addEventListener(`change`,()=>{ne(e.checked)})}),o.querySelectorAll(`[data-action="credentials"]`).forEach(e=>{e.addEventListener(`click`,()=>{R(l.config.isSecure)})}),o.querySelectorAll(`[data-number]`).forEach(e=>{e.addEventListener(`change`,()=>{e.dataset.number===`port`?N(e):e.dataset.number===`maxLevel`?P(e):e.dataset.number===`scratchMaxItems`&&F(e)})}),o.querySelectorAll(`[data-text-setting]`).forEach(e=>{e.addEventListener(`change`,()=>{e.dataset.textSetting===`scratchMaxItemSize`&&I(e)})}),o.querySelector(`[data-action="open-config"]`)?.addEventListener(`click`,()=>{M(async()=>{await _.OpenConfigDir(),m=W(),C()})}),o.querySelector(`.log-filters`)?.addEventListener(`submit`,e=>{e.preventDefault(),M(async()=>{f=J(),u=await _.GetLogs(f),m=W(),C()})}),o.querySelector(`[data-action="refresh-logs"]`)?.addEventListener(`click`,()=>{M(async()=>{f=J(),u=await _.GetLogs(f),m=W(),C()})}),o.querySelector(`[data-action="export-logs"]`)?.addEventListener(`click`,()=>{M(async()=>{f=J();let e=await _.ExportLogs(f);m=e?`已导出到 ${e}`:W(),C()})}),o.querySelector(`[data-action="clear-logs"]`)?.addEventListener(`click`,()=>{window.confirm(`确定清空访问日志？`)&&M(async()=>{await _.ClearLogs(),u=[],m=W(),C()})}),o.querySelectorAll(`[data-action="close-credentials"]`).forEach(e=>{e.addEventListener(`click`,()=>{z()})}),o.querySelector(`.credential-form`)?.addEventListener(`submit`,e=>{e.preventDefault(),L()}),B()}function M(e){e().catch(e=>{m=$(e),C()})}async function ne(e){if(e&&!l.hasCredentials){R(!0);return}M(async()=>{l=await _.UpdateConfig({isSecure:e}),m=W(),C()})}async function N(e){let t=U(e.value,1,65535,`端口`);if(t===null){C();return}if(t!==l.config.port){if(l.running&&!window.confirm(`修改端口需要重启共享服务，是否继续？`)){C();return}M(async()=>{l=await _.UpdateConfig({port:t,restartPort:l.running}),m=W(),C()})}}async function P(e){let t=U(e.value,0,255,`最大访问层级`);if(t===null){C();return}t!==l.config.maxLevel&&M(async()=>{l=await _.UpdateConfig({maxLevel:t}),m=W(),C()})}async function F(e){let t=U(e.value,1,1e4,`临时列表容量`);if(t===null){C();return}t!==l.config.scratchMaxItems&&M(async()=>{l=await _.UpdateConfig({scratchMaxItems:t}),m=W(),C()})}async function I(e){let t=e.value.trim();if(!/^[1-9][0-9]*\s*(B|KB|K|MB|M|GB|G)?$/i.test(t)){m=`单条大小上限格式无效`,C();return}t!==l.config.scratchMaxItemSize&&M(async()=>{l=await _.UpdateConfig({scratchMaxItemSize:t}),m=W(),C()})}async function L(){if(!g)return;let e=H(`username`).trim(),t=H(`password`),n=H(`confirmPassword`),r=g.targetSecure;g={...g,username:e,password:t,confirmPassword:n,error:``};let i=V(e,t,n);if(i){g.error=i,m=``,C();return}M(async()=>{l=await _.SetCredentials({username:e,password:t,confirmPassword:n,enableSecure:r}),g=null,m=W(),C()})}function R(e){g={targetSecure:e,username:``,password:``,confirmPassword:``,error:``},m=``,C()}function z(){g=null,m=W(),C()}function B(){let e=o.querySelector(`.credential-dialog`);if(!e)return;let t=()=>{g&&(g=null,m=W(),C())};e.addEventListener(`cancel`,e=>{e.preventDefault(),t()}),e.addEventListener(`close`,t),e.open||e.showModal(),e.querySelector(`input[name="username"]`)?.focus()}function V(e,t,n){return e?t.trim()?t===n?``:`两次输入的密码不一致`:`密码不能为空`:`用户名不能为空`}function H(e){return o.querySelector(`.credential-form [name="${e}"]`)?.value??``}function U(e,t,n,r){let i=Number(e);return!Number.isInteger(i)||i<t||i>n?(m=`${r}必须在 ${t}-${n} 之间`,null):i}function W(){return h?`浏览器预览模式`:``}function G(){return h?`browser-preview`:`wails-desktop`}function K(e){return e===`browser-preview`?`浏览器预览模式`:`Wails 桌面运行时`}function q(e){return`http://127.0.0.1:${e}`}function J(){let e=o.querySelector(`.log-filters`);if(!e)return f;let t=new FormData(e),n=String(t.get(`event`)??``).trim(),r=Y(String(t.get(`since`)??``)),i=Y(String(t.get(`until`)??``)),a={};return n&&(a.event=n),r&&(a.since=r),i&&(a.until=i),a}function Y(e){if(!e)return null;let t=new Date(e);return Number.isNaN(t.getTime())?null:t.toISOString()}function X(e){if(!e)return``;let t=new Date(e);return Number.isNaN(t.getTime())?``:new Date(t.getTime()-t.getTimezoneOffset()*6e4).toISOString().slice(0,16)}function re(e){let t=new Date(e);return Number.isNaN(t.getTime())?e||`-`:new Intl.DateTimeFormat(void 0,{year:`numeric`,month:`2-digit`,day:`2-digit`,hour:`2-digit`,minute:`2-digit`,second:`2-digit`}).format(t)}function ie(){let e=Date.now(),t=t=>new Date(e-t*6e4).toISOString();return[{time:t(4),clientIP:`192.168.31.18`,method:`GET`,event:`access`,path:`/`,status:200,result:`ok`},{time:t(16),clientIP:`192.168.31.42`,method:`GET`,event:`download`,path:`/photos/2026/spring-trip/very-long-file-name-that-should-wrap-in-the-log-table-without-breaking-layout.jpg`,status:200,result:`sent`},{time:t(28),clientIP:`192.168.31.42`,method:`POST`,event:`upload`,path:`/uploads/report-final.pdf`,status:201,result:`created`},{time:t(44),clientIP:`192.168.31.9`,method:`POST`,event:`login`,path:`/login`,status:200,result:`authenticated`},{time:t(63),clientIP:`192.168.31.77`,method:`GET`,event:`reject`,path:`/private/<script>alert(1)<\/script>.txt`,status:403,result:`blocked`},{time:t(87),clientIP:`192.168.31.51`,method:`GET`,event:`error`,path:`/archive.zip`,status:500,result:`read failed`}]}function ae(e,t){let n=t.event?.trim(),r=Z(t.since),i=Z(t.until);return e.filter(e=>{let t=Z(e.time);return!(n&&e.event!==n||r!==null&&t!==null&&t<r||i!==null&&t!==null&&t>i)})}function Z(e){if(!e)return null;let t=new Date(e).getTime();return Number.isNaN(t)?null:t}function Q(e){return e.replace(/[&<>"']/g,e=>({"&":`&amp;`,"<":`&lt;`,">":`&gt;`,'"':`&quot;`,"'":`&#39;`})[e])}function $(e){return e instanceof Error?e.message:String(e)}