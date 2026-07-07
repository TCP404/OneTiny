<!DOCTYPE html>
<html lang="zh-CN">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="icon" href="/favicon.ico">
    <title>OneTiny - {{- .pathTitle -}}</title>
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
            --amber: #bd7a14;
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
        input {
            font: inherit;
            letter-spacing: 0;
        }

        .shell {
            width: min(1180px, calc(100vw - 32px));
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

        .top-actions {
            display: flex;
            align-items: center;
            justify-content: flex-end;
            gap: 8px;
            flex-wrap: wrap;
        }

        .button,
        .view-button,
        .download-link {
            height: 34px;
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
        .view-button.is-active {
            border-color: var(--brand);
            background: var(--brand);
            color: #fff;
        }

        .button:disabled,
        .button.primary:disabled {
            border-color: var(--line);
            background: #edf2f6;
            color: #95a0ab;
            cursor: not-allowed;
        }

        .view-button {
            min-width: 54px;
        }

        .view-switch {
            display: flex;
            gap: 6px;
            border: 1px solid var(--line);
            background: #edf2f6;
            border-radius: 8px;
            padding: 4px;
        }

        .pathbar,
        .searchbar,
        .upload-panel,
        .content-panel {
            border: 1px solid var(--line);
            border-radius: 8px;
            background: var(--panel);
            box-shadow: var(--shadow);
        }

        .pathbar {
            display: flex;
            align-items: center;
            gap: 8px;
            min-height: 46px;
            margin-bottom: 12px;
            padding: 0 14px;
            overflow: hidden;
            color: var(--muted);
            white-space: nowrap;
        }

        .path-label {
            color: var(--muted);
            font-size: 12px;
            font-weight: 700;
        }

        .path-value {
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            color: var(--ink);
            font-weight: 700;
        }

        .breadcrumbs {
            display: flex;
            align-items: center;
            gap: 8px;
            min-width: 0;
            overflow: hidden;
        }

        .breadcrumb-link,
        .breadcrumb-current {
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            font-weight: 700;
        }

        .breadcrumb-link {
            color: var(--brand-dark);
        }

        .breadcrumb-link:hover,
        .breadcrumb-link:focus {
            text-decoration: underline;
            outline: none;
        }

        .breadcrumb-current {
            color: var(--ink);
        }

        .breadcrumb-separator {
            color: var(--muted);
            flex: 0 0 auto;
        }

        .upload-panel {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 16px;
            margin-bottom: 12px;
            padding: 16px;
            box-shadow: none;
        }

        .upload-copy strong {
            display: block;
            margin-bottom: 5px;
            font-size: 14px;
        }

        .upload-copy span {
            color: var(--muted);
            font-size: 12px;
        }

        .upload-form {
            display: flex;
            align-items: center;
            justify-content: flex-end;
            gap: 8px;
            min-width: min(440px, 100%);
        }

        .file-input {
            position: absolute;
            width: 1px;
            height: 1px;
            overflow: hidden;
            clip: rect(0, 0, 0, 0);
            white-space: nowrap;
        }

        .file-picker-name {
            min-width: 0;
            max-width: 220px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            color: var(--muted);
            font-size: 12px;
        }

        .searchbar {
            display: grid;
            grid-template-columns: minmax(0, 1fr) auto;
            align-items: center;
            gap: 10px;
            margin-bottom: 12px;
            padding: 10px;
            box-shadow: none;
        }

        .search-wrap {
            position: relative;
        }

        .search-input {
            width: 100%;
            height: 38px;
            border: 1px solid var(--line);
            border-radius: 7px;
            background: #fff;
            color: var(--ink);
            outline: none;
            padding: 0 112px 0 12px;
        }

        .search-input:focus {
            border-color: var(--brand);
            box-shadow: 0 0 0 3px rgba(21, 166, 120, 0.14);
        }

        .shortcut {
            position: absolute;
            right: 8px;
            top: 50%;
            transform: translateY(-50%);
            border: 1px solid var(--line);
            border-radius: 5px;
            background: #f7fafc;
            color: var(--muted);
            padding: 3px 7px;
            font-size: 11px;
            font-weight: 700;
            pointer-events: none;
        }

        .count {
            color: var(--muted);
            font-size: 12px;
            font-weight: 700;
            white-space: nowrap;
        }

        .content-panel {
            overflow: hidden;
        }

        .list-view,
        .grid-view {
            display: none;
        }

        .view-list .list-view {
            display: block;
        }

        .view-grid .grid-view {
            display: grid;
        }

        .list-head,
        .list-entry {
            display: grid;
            grid-template-columns: minmax(0, 1fr) 112px 108px;
            align-items: center;
            gap: 14px;
            padding: 0 16px;
        }

        .list-head {
            min-height: 40px;
            background: #f6f8fa;
            color: var(--muted);
            border-bottom: 1px solid var(--line-soft);
            font-size: 11px;
            font-weight: 800;
            text-transform: uppercase;
        }

        .list-entry {
            min-height: 58px;
            border-bottom: 1px solid var(--line-soft);
            cursor: pointer;
        }

        .list-entry:last-child {
            border-bottom: 0;
        }

        .list-entry:hover,
        .list-entry:focus {
            background: #f8fbfa;
            outline: none;
        }

        .file-main {
            display: flex;
            align-items: center;
            gap: 10px;
            min-width: 0;
        }

        .file-icon {
            width: 32px;
            height: 32px;
            border-radius: 7px;
            display: grid;
            place-items: center;
            flex: 0 0 auto;
            font-size: 13px;
            font-weight: 800;
        }

        .file-icon.dir {
            background: #e8f7f1;
            color: var(--brand-dark);
        }

        .file-icon.file {
            background: #e9f0fb;
            color: var(--blue);
        }

        .file-icon.archive {
            background: #f9efe1;
            color: var(--amber);
        }

        .file-name {
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            font-weight: 700;
        }

        .file-kind {
            display: block;
            margin-top: 2px;
            color: var(--muted);
            font-size: 11px;
            font-weight: 500;
        }

        .file-size {
            color: var(--muted);
            text-align: right;
            font-variant-numeric: tabular-nums;
        }

        .file-actions {
            display: flex;
            justify-content: flex-end;
        }

        .download-link {
            height: 30px;
            min-width: 48px;
            padding: 0 10px;
        }

        .download-link:hover,
        .download-link:focus {
            border-color: var(--ink);
            outline: none;
        }

        .grid-view {
            grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
            gap: 12px;
            padding: 14px;
        }

        .grid-card {
            min-height: 142px;
            border: 1px solid var(--line);
            border-radius: 8px;
            background: #fff;
            cursor: pointer;
            padding: 13px;
            display: flex;
            flex-direction: column;
            justify-content: space-between;
            gap: 16px;
        }

        .grid-card:hover,
        .grid-card:focus {
            border-color: rgba(21, 166, 120, 0.55);
            box-shadow: 0 8px 22px rgba(26, 38, 52, 0.08);
            outline: none;
        }

        .grid-top {
            display: flex;
            align-items: flex-start;
            justify-content: space-between;
            gap: 10px;
        }

        .grid-name {
            margin: 10px 0 0;
            font-size: 14px;
            font-weight: 800;
            overflow-wrap: anywhere;
        }

        .grid-meta {
            color: var(--muted);
            font-size: 12px;
            font-variant-numeric: tabular-nums;
        }

        .grid-actions {
            display: flex;
            justify-content: flex-start;
        }

        .empty-state {
            display: none;
            padding: 44px 20px;
            color: var(--muted);
            text-align: center;
            border-top: 1px solid var(--line-soft);
        }

        .has-no-results .empty-state {
            display: block;
        }

        .has-no-results .list-head {
            border-bottom: 0;
        }

        [hidden] {
            display: none !important;
        }

        @media (max-width: 720px) {
            .shell {
                width: min(100vw - 20px, 1180px);
                padding-top: 14px;
            }

            .topbar,
            .upload-panel,
            .upload-form {
                align-items: stretch;
                flex-direction: column;
            }

            .top-actions {
                justify-content: stretch;
            }

            .view-switch,
            .top-actions .button,
            .upload-form .button {
                width: 100%;
            }

            .view-button {
                flex: 1;
            }

            .searchbar {
                grid-template-columns: 1fr;
            }

            .list-head,
            .list-entry {
                grid-template-columns: minmax(0, 1fr) 64px;
                gap: 10px;
                padding: 0 12px;
            }

            .list-head .file-size,
            .list-entry .file-size {
                display: none;
            }

            .list-head .file-actions,
            .list-entry .file-actions {
                display: flex;
            }

            .grid-view {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>

<body>
    <main class="shell view-list" id="fileBrowser">
        <header class="topbar">
            <div class="brand">
                <img class="brand-logo" src="/logo.png" alt="OneTiny">
                <div class="brand-title">
                    <strong>OneTiny</strong>
                    <span>局域网文件访问</span>
                </div>
            </div>
            <div class="top-actions">
                <div class="view-switch" aria-label="切换视图">
                    <button class="view-button is-active" type="button" data-view-toggle="list" aria-pressed="true">列表</button>
                    <button class="view-button" type="button" data-view-toggle="grid" aria-pressed="false">网格</button>
                </div>
            </div>
        </header>

        <section class="pathbar" aria-label="当前路径">
            <span class="path-label">当前目录</span>
            <nav class="breadcrumbs" aria-label="面包屑">
                {{range $i, $crumb := .breadcrumbs}}
                {{if $i}}<span class="breadcrumb-separator" aria-hidden="true">/</span>{{end}}
                {{if $crumb.Current}}
                <span class="breadcrumb-current" aria-current="page">{{$crumb.Name}}</span>
                {{else}}
                <a class="breadcrumb-link" href="{{$crumb.URL}}">{{$crumb.Name}}</a>
                {{end}}
                {{end}}
            </nav>
        </section>

        {{if .upload}}
        <section class="upload-panel" aria-label="上传文件">
            <div class="upload-copy">
                <strong>上传到当前目录</strong>
                <span>选择文件后点击上传，文件会保存到当前目录。</span>
            </div>
            <form class="upload-form" action="/file/upload" method="post" enctype="multipart/form-data" data-upload-form>
                <input type="hidden" name="path" value="{{- .pathTitle -}}">
                <input class="file-input" id="uploadFile" type="file" name="upload_file">
                <span class="file-picker-name" data-file-name>未选择文件</span>
                <label class="button" for="uploadFile">选择文件</label>
                <button class="button primary" type="submit" data-upload-submit disabled>上传</button>
            </form>
        </section>
        {{end}}

        <section class="searchbar" aria-label="页面搜索">
            <div class="search-wrap">
                <input class="search-input" id="fileSearch" type="search" data-shortcut="mod+k" placeholder="搜索当前页面中的文件或目录">
                <span class="shortcut">Ctrl / Cmd + K</span>
            </div>
            <div class="count"><span data-visible-count>{{len .files}}</span> / <span data-total-count>{{len .files}}</span> 项</div>
        </section>

        <section class="content-panel" data-file-list>
            <div class="list-view" aria-label="列表视图">
                <div class="list-head">
                    <div>名称</div>
                    <div class="file-size">大小</div>
                    <div class="file-actions">下载</div>
                </div>

                <div class="list-entry" role="link" tabindex="0" data-open-url="../" data-file-entry data-search-text=".. 返回上一级">
                    <div class="file-main">
                        <span class="file-icon dir" aria-hidden="true">D</span>
                        <span class="file-name">../ 返回上一级</span>
                    </div>
                    <div class="file-size">-</div>
                    <div class="file-actions"></div>
                </div>

                {{range $f := .files}}
                <div class="list-entry" role="link" tabindex="0" data-open-url="{{$f.URLRelPath}}?action=view" data-file-entry data-countable data-search-text="{{$f.Name}}">
                    <div class="file-main">
                        {{if $f.IsDir}}
                        <span class="file-icon dir" aria-hidden="true">D</span>
                        {{else}}
                        <span class="file-icon file" aria-hidden="true">F</span>
                        {{end}}
                        <span class="file-name">{{$f.Name}}</span>
                    </div>
                    <div class="file-size">{{if $f.IsDir}}-{{else}}{{$f.Size}}{{end}}</div>
                    <div class="file-actions">
                        <a class="download-link" href="{{$f.URLRelPath}}?action=dl" download="{{$f.Name}}" data-download-link>下载</a>
                    </div>
                </div>
                {{end}}
            </div>

            <div class="grid-view" aria-label="网格视图">
                <article class="grid-card" role="link" tabindex="0" data-open-url="../" data-file-entry data-search-text=".. 返回上一级">
                    <div>
                        <div class="grid-top">
                            <span class="file-icon dir" aria-hidden="true">D</span>
                            <span class="grid-meta">上级目录</span>
                        </div>
                        <div class="grid-name">../ 返回上一级</div>
                    </div>
                    <div class="grid-actions"></div>
                </article>

                {{range $f := .files}}
                <article class="grid-card" role="link" tabindex="0" data-open-url="{{$f.URLRelPath}}?action=view" data-file-entry data-countable data-search-text="{{$f.Name}}">
                    <div>
                        <div class="grid-top">
                            {{if $f.IsDir}}
                            <span class="file-icon dir" aria-hidden="true">D</span>
                            <span class="grid-meta">目录</span>
                            {{else}}
                            <span class="file-icon file" aria-hidden="true">F</span>
                            <span class="grid-meta">{{$f.Size}}</span>
                            {{end}}
                        </div>
                        <div class="grid-name">{{$f.Name}}</div>
                    </div>
                    <div class="grid-actions">
                        <a class="download-link" href="{{$f.URLRelPath}}?action=dl" download="{{$f.Name}}" data-download-link>下载</a>
                    </div>
                </article>
                {{end}}
            </div>

            <div class="empty-state" data-empty-state>没有匹配的文件或目录</div>
        </section>
    </main>

    <script>
        (function () {
            var shell = document.getElementById("fileBrowser");
            var search = document.getElementById("fileSearch");
            var visibleCount = document.querySelector("[data-visible-count]");
            var totalCount = document.querySelector("[data-total-count]");
            var list = document.querySelector("[data-file-list]");
            var fileInput = document.getElementById("uploadFile");
            var fileName = document.querySelector("[data-file-name]");
            var uploadForm = document.querySelector("[data-upload-form]");
            var uploadSubmit = document.querySelector("[data-upload-submit]");
            var viewButtons = Array.prototype.slice.call(document.querySelectorAll("[data-view-toggle]"));

            function setView(view) {
                var isGrid = view === "grid";
                shell.classList.toggle("view-grid", isGrid);
                shell.classList.toggle("view-list", !isGrid);
                viewButtons.forEach(function (button) {
                    var active = button.getAttribute("data-view-toggle") === view;
                    button.classList.toggle("is-active", active);
                    button.setAttribute("aria-pressed", active ? "true" : "false");
                });
                try {
                    window.localStorage.setItem("onetiny.fileView", view);
                } catch (error) {
                    return;
                }
            }

            function filterEntries() {
                var query = search ? search.value.trim().toLowerCase() : "";
                var entries = Array.prototype.slice.call(document.querySelectorAll("[data-file-entry]"));
                var names = {};
                var allNames = {};

                entries.forEach(function (entry) {
                    var text = (entry.getAttribute("data-search-text") || "").toLowerCase();
                    var visible = query === "" || text.indexOf(query) !== -1;
                    entry.hidden = !visible;
                    if (entry.hasAttribute("data-countable")) {
                        allNames[text] = true;
                    }
                    if (visible && entry.hasAttribute("data-countable")) {
                        names[text] = true;
                    }
                });

                var total = Object.keys(names).length;
                if (visibleCount) {
                    visibleCount.textContent = String(total);
                }
                if (list) {
                    list.classList.toggle("has-no-results", total === 0);
                }
                if (totalCount) {
                    totalCount.textContent = String(Object.keys(allNames).length);
                }
            }

            document.addEventListener("click", function (event) {
                var download = event.target.closest("[data-download-link]");
                if (download) {
                    event.stopPropagation();
                    return;
                }

                var entry = event.target.closest("[data-open-url]");
                if (!entry) {
                    return;
                }
                window.location.href = entry.getAttribute("data-open-url");
            });

            document.addEventListener("keydown", function (event) {
                if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
                    event.preventDefault();
                    if (search) {
                        search.focus();
                        search.select();
                    }
                    return;
                }

                if (event.key !== "Enter") {
                    return;
                }
                var entry = event.target.closest("[data-open-url]");
                if (!entry) {
                    return;
                }
                window.location.href = entry.getAttribute("data-open-url");
            });

            viewButtons.forEach(function (button) {
                button.addEventListener("click", function () {
                    setView(button.getAttribute("data-view-toggle"));
                });
            });

            if (search) {
                search.addEventListener("input", filterEntries);
            }

            if (fileInput && fileName) {
                fileInput.addEventListener("change", function () {
                    var hasFile = fileInput.files && fileInput.files.length > 0;
                    fileName.textContent = hasFile ? fileInput.files[0].name : "未选择文件";
                    if (uploadSubmit) {
                        uploadSubmit.disabled = !hasFile;
                    }
                });
            }

            if (uploadForm) {
                uploadForm.addEventListener("submit", function (event) {
                    if (!fileInput || !fileInput.files || fileInput.files.length === 0) {
                        event.preventDefault();
                    }
                });
            }

            try {
                setView(window.localStorage.getItem("onetiny.fileView") || "list");
            } catch (error) {
                setView("list");
            }
            filterEntries();
        })();
    </script>
</body>

</html>
