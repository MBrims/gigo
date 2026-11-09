import { useState } from "react";

function GrabQueryInfo() {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get('file');
}

function Query() {
    const queryInfo = GrabQueryInfo();
    return (
        <h1>{queryInfo ? "Found query: " + queryInfo : "No query found."}</h1>
    );
}

function HashDisplay() {
    const hash = window.location.hash;
    if (hash.length == 0) {
        return <h1>No hash found.</h1>;
    }

    // Remove the leading #
    return <p>Found hash: {hash.slice(1)}</p>
}

const Download = () => {
    const [loadedQuery, setLoadedQuery] = useState<boolean>(false);
    const [loadedData, setLoadedData] = useState<string>();

    if (!loadedQuery) {
        const queryInfo = GrabQueryInfo();
        fetch("/download?file=" + queryInfo).then((r) => {
            if (r.ok) {
                // Split it and then join it back together.
                const suggestedFileName = r.headers.get("content-disposition")?.split("=").slice(1).join("=");
                console.log(suggestedFileName);

                r.text().then((v) => {
                    setLoadedData(v);
                    setLoadedQuery(true);
                });
            }
        });
    }

    if (loadedQuery) {
        return (
            <div>
                <p>{loadedData}</p>
                <HashDisplay />
                <a href="/">Back to Home</a>
            </div>
        )
    }

    return (
        <div>
            <Query />
            <HashDisplay />
            <a href="/">Back to Home</a>
        </div>
    );
};

export default Download;
