/**
 * This page should just have some stats about the running server; total space taken up sorta deal, # of files, etc.
 */

import { useEffect, useState, type JSX } from "react"
import { FormatBytes } from "../helpers";

function statusOk(status: number) {
    return (status >= 200 && status < 300)
}

type statsResponseT = { status: number, resp: unknown }

// Returns a statsResponseT, if .status is OK, then resp will be an object representing the JSON.
async function requestStats(): Promise<statsResponseT> {
    const resp = await fetch("/statistics")
    if (resp.ok) {
        const json = await resp.json();
        console.log(json);

        return { status: resp.status, resp: json }
    }

    return { status: resp.status, resp: resp.statusText }
}



const Stats = () => {
    const [loadedStats, setLoadedStats] = useState<object>();

    // We only want to run this once.
    useEffect(() => {
        requestStats().then(r => {
            if (statusOk(r.status)) {
                setLoadedStats(r.resp as object);
            }
        });
    }, []);

    const statValues: JSX.Element[] = []

    if (loadedStats) {

        for (const key of Object.keys(loadedStats)) {
            const lowerKey = key.toLowerCase();

            //@ts-expect-error We're grabbing the keys already.
            let value = loadedStats[key].toString();

            switch (lowerKey) {
                case "total":
                case "allotted":
                    value = FormatBytes(Number.parseInt(value));
                    break;
            }

            statValues.push(<li>{key} : {value}</li>)
        }
    }

    return (
        <>
            <h1>Stats</h1>
            <p>Some statistics about the server. Currently it's just slapping the JSON down as is.</p>
            <hr />
            <ul>
                {statValues}
            </ul>
        </>
    )
}

export default Stats;