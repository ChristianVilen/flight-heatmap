<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import L, { Map as LeafletMap, type HeatLayer } from "leaflet";
  import "leaflet/dist/leaflet.css";
  import "leaflet.heat";

  type HeatPoint = {
    lat: number;
    lon: number;
    count: number; // intensity
  };

  const POLL_INTERVAL_MS = 10000;
  const helsinkiAirportCoords: [number, number] = [60.3172, 24.9633];
  const minuteOptions = [
    { label: "5 minutes", value: 5 },
    { label: "15 minutes", value: 15 },
    { label: "30 minutes", value: 30 },
    { label: "1 hour", value: 60 },
    { label: "4 hours", value: 240 },
    { label: "No limit", value: null },
  ];

  let map: LeafletMap;
  let heatLayer: HeatLayer;
  let previousBin = 0;
  let selectedMinutes = 30;
  let interval: ReturnType<typeof setInterval>;

  function getBin(zoom: number): number {
    return zoom >= 13 ? 160 : zoom >= 11 ? 80 : 40;
  }

  async function fetchHeatData(
    bin: number,
  ): Promise<[number, number, number][]> {
    const url = new URL("/api/heatmap", window.location.origin);
    url.searchParams.set("bin", bin.toString());
    if (selectedMinutes !== null) {
      url.searchParams.set("minutes", selectedMinutes.toString());
    }

    const res = await fetch(url.toString());
    const points: HeatPoint[] = await res.json();

    if (!points) return [];
    return points.map((p) => [p.lat, p.lon, p.count]);
  }

  async function updateHeatmap() {
    if (!map) return;
    const bin = getBin(map.getZoom());
    const data = await fetchHeatData(bin);
    heatLayer.setLatLngs(data);
  }

  function setupZoomHandler() {
    map.on("zoomend", async () => {
      if (!map) return;

      const currentBin = getBin(map.getZoom());
      if (currentBin !== previousBin) {
        previousBin = currentBin;
        await updateHeatmap();
      }
    });
  }

  onMount(async () => {
    map = L.map("map", { preferCanvas: true }).setView(
      helsinkiAirportCoords,
      12,
    );

    L.tileLayer(
      "https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png",
      {
        attribution: `&copy; <a href="https://www.openstreetmap.org/copyright" target="_blank">OpenStreetMap</a>,
        &copy; <a href="https://carto.com/attributions" target="_blank">CARTO</a>`,
        subdomains: "abcd",
        maxZoom: 14,
      },
    ).addTo(map);

    previousBin = getBin(map.getZoom());
    const initialData = await fetchHeatData(previousBin);
    heatLayer = L.heatLayer(initialData, {
      radius: 15,
      blur: 10,
      maxZoom: 14,
    }).addTo(map);

    setupZoomHandler();
    interval = setInterval(updateHeatmap, POLL_INTERVAL_MS);
  });

  onDestroy(() => {
    clearInterval(interval);
    map?.remove();
  });

  function resizeMap() {
    map?.invalidateSize();
  }
</script>

<svelte:window on:resize={resizeMap} />

<div id="map" class="w-full h-full">
  <div
    class="absolute top-4 right-4 z-[1000] bg-white shadow-md rounded px-3 py-1"
  >
    <label for="minutes" class="mr-2 font-semibold">Minutes:</label>
    <select
      id="minutes"
      bind:value={selectedMinutes}
      class="border px-2 py-1 rounded"
      on:change={() => updateHeatmap()}
    >
      {#each minuteOptions as option}
        <option value={option.value}>{option.label}</option>
      {/each}
    </select>
  </div>
</div>
