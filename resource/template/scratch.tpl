<!DOCTYPE html>
<html lang="zh-CN">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="icon" href="/favicon.ico">
    <title>OneTiny - 临时列表</title>
    <style>
        :root {
            color-scheme: light;
            --bg: #f4f7f9;
            --panel: #ffffff;
            --ink: #17202a;
            --muted: #68717d;
            --line: #d9e0e7;
            --line-soft: #e7edf2;
            --brand: #15a678;
            --brand-dark: #0f7f5d;
            --blue: #2e6fd3;
            --danger: #b83d3d;
            --shadow: 0 12px 30px rgba(26, 38, 52, 0.08);
        }

        * {
            box-sizing: border-box;
        }

        body {
            margin: 0;
            min-height: 100vh;
            background: var(--bg);
            color: var(--ink);
            font-family: Inter, "SF Pro Display", "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
            font-size: 14px;
            line-height: 1.5;
            letter-spacing: 0;
            -webkit-font-smoothing: antialiased;
        }

        a {
            color: inherit;
            text-decoration: none;
        }

        button,
        input,
        textarea {
            font: inherit;
            letter-spacing: 0;
        }

        .shell {
            width: min(1080px, calc(100vw - 32px));
            margin: 0 auto;
            padding: 28px 0 40px;
        }

        .topbar {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 16px;
            margin-bottom: 18px;
        }

        .brand {
            display: flex;
            align-items: center;
            gap: 12px;
            min-width: 0;
        }

        .brand-logo {
            width: 48px;
            height: 48px;
            border-radius: 8px;
            flex: 0 0 auto;
            display: block;
            background: #fff;
            object-fit: contain;
            box-shadow: 0 1px 3px rgba(26, 38, 52, 0.08);
        }

        .brand-title {
            min-width: 0;
        }

        .brand-title strong {
            display: block;
            font-size: 18px;
            line-height: 1.2;
        }

        .brand-title span {
            display: block;
            margin-top: 4px;
            color: var(--muted);
            font-size: 12px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .top-actions,
        .form-actions,
        .item-actions {
            display: flex;
            align-items: center;
            justify-content: flex-end;
            gap: 8px;
            flex-wrap: wrap;
        }

        .button,
        .nav-button {
            min-height: 34px;
            border: 1px solid var(--line);
            border-radius: 6px;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            gap: 7px;
            background: var(--panel);
            color: var(--ink);
            padding: 0 12px;
            font-size: 12px;
            font-weight: 700;
            white-space: nowrap;
            cursor: pointer;
        }

        .button.primary,
        .nav-button.is-active {
            border-color: var(--brand);
            background: var(--brand);
            color: #fff;
        }

        .button:hover,
        .button:focus,
        .nav-button:hover,
        .nav-button:focus {
            border-color: var(--brand);
            outline: none;
        }

        .panel,
        .item-card,
        .error {
            border: 1px solid var(--line);
            border-radius: 8px;
            background: var(--panel);
            box-shadow: var(--shadow);
        }

        .error {
            margin-bottom: 12px;
            border-color: rgba(184, 61, 61, 0.32);
            background: #fff7f7;
            color: var(--danger);
            padding: 12px 14px;
            font-weight: 700;
        }

        .composer {
            display: grid;
            grid-template-columns: minmax(0, 1fr) minmax(280px, 360px);
            gap: 12px;
            margin-bottom: 16px;
        }

        .panel {
            padding: 16px;
            box-shadow: none;
        }

        .panel h1,
        .panel h2 {
            margin: 0 0 10px;
            font-size: 16px;
            line-height: 1.3;
        }

        .panel-meta,
        .item-meta,
        .empty-state {
            color: var(--muted);
            font-size: 12px;
        }

        .field {
            width: 100%;
            border: 1px solid var(--line);
            border-radius: 7px;
            background: #fff;
            color: var(--ink);
            outline: none;
        }

        .field:focus {
            border-color: var(--brand);
            box-shadow: 0 0 0 3px rgba(21, 166, 120, 0.14);
        }

        textarea.field {
            min-height: 144px;
            resize: vertical;
            padding: 10px 12px;
            overflow-wrap: anywhere;
        }

        .file-field {
            min-height: 42px;
            padding: 9px 10px;
        }

        .form-actions {
            margin-top: 10px;
        }

        .list-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
            margin: 18px 0 10px;
        }

        .list-header h2 {
            margin: 0;
            font-size: 16px;
        }

        .items {
            display: grid;
            gap: 12px;
        }

        .item-card {
            padding: 14px;
            box-shadow: none;
        }

        .item-head {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
            margin-bottom: 10px;
        }

        .item-title {
            min-width: 0;
            display: flex;
            align-items: center;
            gap: 8px;
            font-weight: 800;
        }

        .item-badge {
            flex: 0 0 auto;
            border-radius: 6px;
            background: #e8f7f1;
            color: var(--brand-dark);
            padding: 4px 7px;
            font-size: 11px;
            font-weight: 800;
        }

        .item-id {
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            color: var(--muted);
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
            font-size: 12px;
        }

        .text-preview {
            width: 100%;
            min-height: 112px;
            border: 1px solid var(--line-soft);
            border-radius: 7px;
            background: #f9fbfc;
            color: var(--ink);
            padding: 10px 12px;
            resize: vertical;
            overflow: auto;
            overflow-wrap: anywhere;
            white-space: pre-wrap;
        }

        .image-preview {
            display: block;
            max-width: 100%;
            max-height: 460px;
            border: 1px solid var(--line-soft);
            border-radius: 8px;
            background: #f9fbfc;
            object-fit: contain;
        }

        .empty-state {
            border: 1px dashed var(--line);
            border-radius: 8px;
            background: rgba(255, 255, 255, 0.7);
            padding: 36px 18px;
            text-align: center;
        }

        [data-copy-done] {
            color: var(--brand-dark);
        }

        @media (max-width: 760px) {
            .shell {
                width: min(100vw - 20px, 1080px);
                padding-top: 14px;
            }

            .topbar,
            .composer,
            .item-head {
                align-items: stretch;
                grid-template-columns: 1fr;
                flex-direction: column;
            }

            .top-actions,
            .form-actions,
            .item-actions {
                justify-content: stretch;
            }

            .top-actions .nav-button,
            .form-actions .button,
            .item-actions .button {
                width: 100%;
            }
        }
    </style>
</head>

<body>
    <main class="shell" id="scratchPage">
        <header class="topbar">
            <div class="brand">
                <img class="brand-logo" src="/logo.png" alt="OneTiny">
                <div class="brand-title">
                    <strong>OneTiny</strong>
                    <span>局域网临时文本与图片</span>
                </div>
            </div>
            <nav class="top-actions" aria-label="主导航">
                <a class="nav-button" href="/file/">文件共享</a>
                <a class="nav-button is-active" href="/scratch/" aria-current="page">临时列表</a>
            </nav>
        </header>

        {{if .error}}
        <div class="error" role="alert">{{.error}}</div>
        {{end}}

        <section class="composer" aria-label="新增临时内容">
            <form class="panel" action="/scratch/items" method="post" data-text-form>
                <h1>添加文本</h1>
                <input type="hidden" name="kind" value="text">
                <textarea class="field" name="text" placeholder="粘贴或输入要临时共享的文本"></textarea>
                <div class="form-actions">
                    <button class="button primary" type="submit">添加文本</button>
                </div>
            </form>

            <form class="panel" action="/scratch/items" method="post" enctype="multipart/form-data" data-image-form>
                <h2>添加图片</h2>
                <p class="panel-meta">支持 PNG、JPEG、GIF、WebP。容量限制：{{.limits.MaxItemBytes}} B，最多保留 {{.limits.MaxItems}} 项。</p>
                <input type="hidden" name="kind" value="image">
                <input class="field file-field" type="file" name="image" accept="image/png,image/jpeg,image/gif,image/webp">
                <div class="form-actions">
                    <button class="button primary" type="submit">上传图片</button>
                </div>
            </form>
        </section>

        <section aria-label="临时内容列表">
            <div class="list-header">
                <h2>临时内容</h2>
                <div class="panel-meta">{{len .items}} / {{.limits.MaxItems}} 项</div>
            </div>

            {{if .items}}
            <div class="items">
                {{range $item := .items}}
                <article class="item-card">
                    <div class="item-head">
                        <div class="item-title">
                            {{if eq $item.Kind "image"}}
                            <span class="item-badge">图片</span>
                            {{else}}
                            <span class="item-badge">文本</span>
                            {{end}}
                            <span class="item-id">{{$item.ID}}</span>
                        </div>
                        <div class="item-meta">{{$item.Size}} B</div>
                    </div>

                    {{if eq $item.Kind "image"}}
                    <img class="image-preview" src="/scratch/items/{{$item.ID}}" alt="临时图片">
                    <div class="item-actions" style="margin-top: 10px;">
                        <a class="button" href="/scratch/items/{{$item.ID}}?download=1">下载图片</a>
                    </div>
                    {{else}}
                    <textarea class="text-preview" readonly data-copy-source>{{printf "%s" $item.Preview}}</textarea>
                    <div class="item-actions" style="margin-top: 10px;">
                        <button class="button" type="button" data-copy-button data-copy-url="/scratch/items/{{$item.ID}}">复制文本</button>
                        <a class="button" href="/scratch/items/{{$item.ID}}?download=1">下载文本</a>
                    </div>
                    {{end}}
                </article>
                {{end}}
            </div>
            {{else}}
            <div class="empty-state">暂无临时内容</div>
            {{end}}
        </section>
    </main>

    <script>
        (function () {
            function postForm(formData) {
                return window.fetch("/scratch/items", {
                    method: "POST",
                    headers: { "Accept": "application/json" },
                    body: formData
                }).then(function (response) {
                    if (!response.ok) {
                        return response.json().catch(function () {
                            return {};
                        }).then(function (payload) {
                            throw new Error(payload.error || "添加失败");
                        });
                    }
                    window.location.reload();
                });
            }

            function postText(text) {
                var formData = new window.FormData();
                formData.append("kind", "text");
                formData.append("text", text);
                return postForm(formData);
            }

            function postImage(file) {
                var formData = new window.FormData();
                formData.append("kind", "image");
                formData.append("image", file, file.name || "clipboard-image");
                return postForm(formData);
            }

            function showSubmitError(error) {
                window.alert((error && error.message) || "添加失败");
            }

            function copyTextFallback(text) {
                return new Promise(function (resolve, reject) {
                    var textarea = document.createElement("textarea");
                    var active = document.activeElement;
                    textarea.value = text;
                    textarea.setAttribute("readonly", "");
                    textarea.style.position = "fixed";
                    textarea.style.left = "-9999px";
                    textarea.style.top = "0";
                    textarea.style.opacity = "0";
                    document.body.appendChild(textarea);
                    textarea.focus();
                    textarea.select();
                    textarea.setSelectionRange(0, textarea.value.length);
                    try {
                        if (!document.execCommand("copy")) {
                            throw new Error("copy command failed");
                        }
                        resolve();
                    } catch (error) {
                        reject(error);
                    } finally {
                        document.body.removeChild(textarea);
                        if (active && active.focus) {
                            active.focus();
                        }
                    }
                });
            }

            function copyText(text) {
                if (window.navigator && window.navigator.clipboard && window.navigator.clipboard.writeText) {
                    return window.navigator.clipboard.writeText(text).catch(function () {
                        return copyTextFallback(text);
                    });
                }
                return copyTextFallback(text);
            }

            function isEditableFocused() {
                var active = document.activeElement;
                if (!active) {
                    return false;
                }
                if (active.isContentEditable) {
                    return true;
                }
                var editable = active.closest ? active.closest("[contenteditable]") : null;
                if (editable && editable.getAttribute("contenteditable") !== "false") {
                    return true;
                }
                var tag = active.tagName;
                return tag === "TEXTAREA" || tag === "INPUT" || tag === "SELECT";
            }

            document.addEventListener("paste", function (event) {
                var clipboard = event.clipboardData;
                if (!clipboard) {
                    return;
                }
                if (isEditableFocused()) {
                    return;
                }

                var items = Array.prototype.slice.call(clipboard.items || []);
                for (var i = 0; i < items.length; i += 1) {
                    if (items[i].type && items[i].type.indexOf("image/") === 0) {
                        var file = items[i].getAsFile();
                        if (file) {
                            event.preventDefault();
                            postImage(file).catch(showSubmitError);
                            return;
                        }
                    }
                }

                var text = clipboard.getData("text/plain");
                if (text) {
                    event.preventDefault();
                    postText(text).catch(showSubmitError);
                }
            });

            document.addEventListener("click", function (event) {
                var button = event.target.closest("[data-copy-button]");
                if (!button) {
                    return;
                }
                var copyUrl = button.getAttribute("data-copy-url");
                var card = button.closest(".item-card");
                var source = card ? card.querySelector("[data-copy-source]") : null;
                var textPromise = copyUrl
                    ? window.fetch(copyUrl).then(function (response) {
                        if (!response.ok) {
                            throw new Error("复制失败");
                        }
                        return response.text();
                    })
                    : Promise.resolve(source ? source.value : "");
                textPromise.then(function (text) {
                    return copyText(text);
                }).then(function () {
                    button.setAttribute("data-copy-done", "true");
                    button.textContent = "已复制";
                    window.setTimeout(function () {
                        button.removeAttribute("data-copy-done");
                        button.textContent = "复制文本";
                    }, 1400);
                }).catch(function () {
                    window.alert("复制失败");
                });
            });
        })();
    </script>
</body>

</html>
