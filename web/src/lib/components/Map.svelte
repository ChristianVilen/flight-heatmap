<script lang="ts">
  import L, { bounds, Map as LeafletMap } from "leaflet";
  import "leaflet/dist/leaflet.css";
  import "leaflet.heat";

  type HeatPoint = {
    lat_bin: number;
    lon_bin: number;
    count: number;
  };

  let map: LeafletMap | null = null;

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

  async function fetchAndRenderHeat() {
    const res = await fetch("/api/heatmap");
    const points: HeatPoint[] = await res.json();
    const heatData = points.map((p) => [p.LatBin, p.LonBin, p.Count]);

    L.heatLayer(heatData, {
      radius: 15,
      blur: 10,
      maxZoom: 14,
      bounds: map?.getBounds(),
    }).addTo(map);
  }

  function mapAction(container: HTMLElement) {
    map = createMap(container);

    fetchAndRenderHeat();

    return {
      destroy: () => {
        map?.remove();
        map = null;
      },
    };
  }

  function resizeMap() {
    map?.invalidateSize();
  }
</script>

<svelte:window on:resize={resizeMap} />

<div use:mapAction class="w-full h-full"></div>

<style>
</style>
