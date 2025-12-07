// Initialize CodeMirror
var editor = CodeMirror.fromTextArea(document.getElementById("code-editor"), {
    lineNumbers: true,
    mode: "go",
    theme: "monokai"
});

// WebSocket connection
// WebSocket connection
var pathParts = window.location.pathname.split('/');
var interviewId = pathParts[pathParts.length - 1];
var wsUrl = "ws://" + window.location.host + "/ws/" + interviewId;

// If we are on the root path, don't connect yet (or maybe connect to a default?)
// Actually, the plan says we redirect to /interview/{id}, so we should be good.
// But let's handle the case where we might be on root.
if (window.location.pathname === '/' || window.location.pathname === '/index.html') {
    // We are on the landing page, no websocket needed yet.
    console.log("On landing page, waiting for interview creation.");
} else {
    var socket = new WebSocket(wsUrl);

    socket.onopen = function () {
        console.log("Connected to WebSocket for session: " + interviewId);
    };

    socket.onmessage = function (event) {
        var data = JSON.parse(event.data);

        // If we receive code update
        if (data.type === 'code_update') {
            // Only update if content is different to avoid cursor jumping issues
            // In a real app we'd use OT/CRDT, but for bootstrap we just replace
            // if it's not the one we just sent (naive check)
            if (editor.getValue() !== data.content) {
                var cursor = editor.getCursor();
                editor.setValue(data.content);
                editor.setCursor(cursor);
            }
        }
    };

    socket.onclose = function () {
        console.log("Disconnected from WebSocket");
    };

    // Listen for changes in the editor
    editor.on("change", function (cm, change) {
        if (change.origin !== "setValue") {
            var content = cm.getValue();
            if (socket.readyState === WebSocket.OPEN) {
                socket.send(JSON.stringify({
                    type: 'code_update',
                    content: content
                }));
            }
        }
    });
}
