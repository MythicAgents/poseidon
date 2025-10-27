function(task, responses) {
    if (task.status.includes("error")) {
        const combined = responses.reduce((prev, cur) => prev + cur, "");
        return { 'plaintext': combined };
    }

    if (task.completed) {
        if (responses.length > 0) {
            let mediaArray = [];
            let filename_pieces = task.display_params.split("/");

            for (let i = 0; i < responses.length; i++) {
                try {
                    let data = JSON.parse(responses[i]);
                    // Try to get filename, fallback to display_params with index
                    let filename = data["filename"] 
                        ? data["filename"] 
                        : `${filename_pieces[filename_pieces.length - 1]}_${i+1}`;

                    if (data["file_id"]) {
                        mediaArray.push({
                            "filename": filename,
                            "agent_file_id": data["file_id"],
                        });
                    }
                } catch (err) {
                    // Ignore or log parse errors for any non-file responses
                }
            }

            if (mediaArray.length > 0) {
                return { "media": mediaArray };
            } else {
                return { "plaintext": "No files found in responses." };
            }
        } else {
            return { "plaintext": "No data to display..." };
        }
    } else {
        // Show progress while the task is running
        if (responses.length > 0) {
            try {
                const task_data = JSON.parse(responses[0]);
                let message = "Downloading file(s)";
                if (task_data["total_chunks"]) {
                    message += " with " + task_data["total_chunks"] + " total chunks...";
                } else {
                    message += "...";
                }
                if (responses.length > 1) {
                    message += "\n" + responses[responses.length - 1];
                }
                return { "plaintext": message };
            } catch (e) {
                return { "plaintext": "Awaiting file download progress..." };
            }
        }
        return { "plaintext": "No data yet..." };
    }
}
