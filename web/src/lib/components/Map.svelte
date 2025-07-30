<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import L, { Map as LeafletMap, type HeatLayer } from "leaflet";
  import "leaflet/dist/leaflet.css";
  import "leaflet.heat";

  type MarkerData = {
    id: number;
    lat: number;
    lon: number;
    count: number;
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
  let markerLayerGroup: L.LayerGroup = L.layerGroup();

  let currentMode: "heatmap" | "markers" = "heatmap";
  const ZOOM_THRESHOLD = 13;

  function getBin(zoom: number): number {
    return zoom >= 13 ? 160 : zoom >= 11 ? 80 : 40;
  }

  async function fetchMarkerData(bin: number): Promise<MarkerData[]> {
    const url = new URL("/api/heatmap", window.location.origin);
    url.searchParams.set("bin", bin.toString());
    if (selectedMinutes !== null) {
      url.searchParams.set("minutes", selectedMinutes.toString());
    }

    const res = await fetch(url.toString());
    const points: MarkerData[] = await res.json();

    if (!points) return [];
    return points;
  }

  async function updateHeatmap() {
    if (!map || currentMode !== "heatmap") return;

    const bin = getBin(map.getZoom());
    const data = await fetchMarkerData(bin);

    if (heatLayer) {
      heatLayer.setLatLngs(data.map((p) => [p.lat, p.lon, p.count]));
    }
  }

  function setupZoomHandler() {
    map.on("zoomend", async () => {
      const zoom = map.getZoom();
      const currentBin = getBin(zoom);

      if (zoom >= ZOOM_THRESHOLD && currentMode !== "markers") {
        currentMode = "markers";

        if (heatLayer) map.removeLayer(heatLayer);
        previousBin = currentBin;

        const data = await fetchMarkerData(currentBin);
        renderAircraftMarkers(data);
      } else if (zoom < ZOOM_THRESHOLD && currentMode !== "heatmap") {
        currentMode = "heatmap";

        markerLayerGroup.clearLayers();
        previousBin = currentBin;

        const data = await fetchMarkerData(currentBin);
        heatLayer = L.heatLayer(
          data.map((p) => [p.lat, p.lon, p.count]),
          {
            radius: 15,
            blur: 10,
            maxZoom: 14,
          },
        ).addTo(map);
      } else if (currentBin !== previousBin) {
        previousBin = currentBin;

        const data = await fetchMarkerData(currentBin);
        if (currentMode === "heatmap" && heatLayer) {
          heatLayer.setLatLngs(data.map((p) => [p.lat, p.lon, p.count]));
        } else {
          renderAircraftMarkers(data);
        }
      }
    });
  }

  function renderAircraftMarkers(data: MarkerData[]) {
    markerLayerGroup.clearLayers();

    data.forEach((plane) => {
      if (plane.lat && plane.lon) {
        const marker = L.marker([plane.lat, plane.lon])
          // .bindPopup(
          //   `<strong>${plane.Callsign || "?"}</strong><br>${plane.OriginCountry || "Unknown"}`,
          // )
          .on("click", () => {
            console.log("Clicked ID:", plane.id);
          });

        markerLayerGroup.addLayer(marker);
      }
    });

    markerLayerGroup.addTo(map!);
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

    const initialBin = getBin(map.getZoom());
    previousBin = initialBin;

    const initialData = await fetchMarkerData(initialBin);

    if (map.getZoom() < ZOOM_THRESHOLD) {
      currentMode = "heatmap";
      heatLayer = L.heatLayer(
        initialData.map((p) => [p.lat, p.lon, p.count]),
        {
          radius: 15,
          blur: 10,
          maxZoom: 14,
        },
      ).addTo(map);
    } else {
      currentMode = "markers";
      renderAircraftMarkers(initialData);
    }

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
