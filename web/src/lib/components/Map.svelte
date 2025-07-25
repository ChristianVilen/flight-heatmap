<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import L, { Map as LeafletMap, type HeatLayer } from "leaflet";
  import "leaflet/dist/leaflet.css";
  import "leaflet.heat";

  type HeatPoint = {
    LatBin: number;
    LonBin: number;
    Count: number; // intensity
  };

  const POLL_INTERVAL_10_SEC = 10000;

  let map: LeafletMap | null = null;
  let heatLayer: HeatLayer;

  const helsinkiAirportCoords: [number, number] = [60.3172, 24.9633];

  function createMap(container: HTMLElement): LeafletMap {
    const m = L.map(container, { preferCanvas: true }).setView(
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
    ).addTo(m);

    return m;
  }

  async function fetchHeatData(): Promise<[number, number, number][]> {
    const res = await fetch("/api/heatmap");
    const points: HeatPoint[] = await res.json();

    return points.map((p) => [p.LatBin, p.LonBin, p.Count]);
  }

  async function updateHeatmap() {
    const heatData = await fetchHeatData();
    if (heatLayer) {
      heatLayer.setLatLngs(heatData);
    }
  }

  let interval: ReturnType<typeof setInterval>;

  onMount(async () => {
    map = createMap(document.getElementById("map")!);

    const initialData = await fetchHeatData();
    heatLayer = L.heatLayer(initialData, {
      radius: 15,
      blur: 10,
      maxZoom: 14,
    }).addTo(map);

    interval = setInterval(updateHeatmap, POLL_INTERVAL_10_SEC);
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

<div id="map" class="w-full h-full"></div>
