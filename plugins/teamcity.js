function init() {
    let tableStatement = "CREATE TABLE IF NOT EXISTS teamcity (teamcity_id INTEGER PRIMARY KEY AUTOINCREMENT, room_id TEXT NOT NULL,room_hash TEXT NOT NULL, build_id TEXT NOT NULL);";
    try {
        DBPS(tableStatement);
    } catch (e) {
        LogError(e);
    }
    AddRoute("/teamcity/webhook/{hash}", "POST", function callback(data) {
        //LogInfo(JSON.stringify(data));
        let res = Object.assign({}, data);

        let matrixRoom = "";
        let matrixMessage = "";
        let roomBuildIDS = [];

        try {
            let rows = DBQuery("SELECT room_id, build_id FROM teamcity WHERE room_hash = ?", "string", data.params.hash);
            if (rows.length === 0) {
                return res;
            }
            for (row of rows) {
                matrixRoom = row[0];
                roomBuildIDS.push(row[1]);
            }
        } catch (e) {
            LogError("Whhops. Couldn't get data from database during http request");
            return res;
        }

        const body = data.body.build || {buildTypeId: ""};
        const buildID = body.buildTypeId || "";

        if (!roomBuildIDS.includes(buildID.toLowerCase())) {
            LogInfo("room is not listening for this repository");
            return res;
        }

        const buildResult = body.buildResult || "";

        if (buildResult === "") {
            LogInfo("unsupported event");
            return res;
        }

        const buildFullName = body.buildFullName || "";
        const buildStatusUrl = body.buildStatusUrl || "";
        const triggeredBy = body.triggeredBy || "";

        matrixMessage += "[" + buildFullName + "](" + buildStatusUrl + ") ";
        // Test Projekt 123 / BuildConfigTest has finished. Status: success
        if (buildResult === "running") {        // Started
            matrixMessage += "has been started ⚙";
        } else if (buildResult === "success") {
            matrixMessage += "has finished with status <font color=\"#00FF00\">**success**</font> 🎉";
        } else if (buildResult === "failed") {
            matrixMessage += "has <font color=\"#FF0000\">**failed**</font> ❌";
        } else if (buildResult === "interrupted") {
            matrixMessage += "has been interrupted ‼";
        } else {
            matrixMessage += "Status: " + (buildResult.charAt(0).toUpperCase() + buildResult.slice(1));
        }

        if (triggeredBy !== "") {
            matrixMessage += "\nTriggered by: " + triggeredBy;
        }

        SendMessage(matrixRoom, matrixMessage);

        return res;
    });
}

function onMessage(data) {
    let res = Object.assign({}, data);
    if (!res.message.startsWith("!teamcity") && !res.message.startsWith("!tc")) return;
    let args = res.message.split(" ");
    if (args.length <= 1) {
        res.response = "Hey there 👋🏼";
        return res
    }
    switch (args[1]) {
        case "test":
        case "ping":
            res.response = "Hey 👋 I'm here to respond to your webhooks of teamcity. 🧃";
            break;
        case "list":
            res.response = "I listen on the following build configurations 🦻  \n";
            try {
                let rows = DBQuery("SELECT * FROM teamcity WHERE room_id = ?", "string", data.roomID);
                if (rows.length === 0) {
                    res.response = "**I listen no none of your build configurations** 😴";
                    break;
                }

                for (row of rows) {
                    res.response += "- " + row[3] + "\n";
                }
            } catch (e) {
                LogError("Whhops. Couldn't get data from database");
                res.response = "Something went wrong while fetching your data from the database 😭";
                break;
            }
            break;
        case "webhook":
            res.response = "Please use [this link](" + this.webServer.baseURL + "/teamcity/webhook/" + SHA256(data.roomID) + ") as your webhook URL and use *Legacy Webhook (JSON)* as type";
            break;
        case "add": {
            if (args.length !== 3) {
                res.response = "Huh? Have you missed to add the build id as argument?";
                break;
            }

            const buildID = args[2].toLowerCase();

            try {
                DBPS("INSERT INTO teamcity(room_id, room_hash, build_id) VALUES (?, ?, ?)",
                    "string", data.roomID, "string", SHA256(data.roomID), "string", buildID);
            } catch (e) {
                LogError("Whhops. Couldn't insert data into database");
                res.response = "Something went wrong while adding your build configuration to my watchlist 😭";
                break;
            }

            res.response = "I added *" + buildID + "* to my watchlist ✅";
            break;
        }

        case "rm":
        case "remove": {
            if (args.length !== 3) {
                res.response = "Huh? Have you missed to add the build id as argument?";
                break;
            }

            const buildID = args[2].toLowerCase();

            try {
                DBPS("DELETE FROM teamcity WHERE room_id = ? AND build_id = ?", "string", data.roomID, "string", buildID);
            } catch (e) {
                LogError("Whhops. Couldn't delete data from database");
                res.response = "Something went wrong while removing your build configuration from my watchlist 😭";
                break;
            }

            res.response = "I removed *" + buildID + "* from my watchlist 😪";
            break;
        }

    }
    return res;
}