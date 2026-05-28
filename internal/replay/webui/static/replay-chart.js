(function () {
	const density = JSON.parse(document.getElementById("density-data").textContent);
	const phases = JSON.parse(document.getElementById("phase-data").textContent);
	const engs = JSON.parse(document.getElementById("eng-data").textContent);

	const traces = density.map((d) => ({
		name: d.name,
		type: "bar",
		x: d.x,
		y: d.y,
	}));

	const shapes = phases.map((p) => ({
		type: "line", xref: "x", yref: "paper",
		x0: p.x, x1: p.x, y0: 0, y1: 1,
		line: { dash: "dash", width: 1, color: "#555" },
	})).concat(engs.map((e) => ({
		type: "line", xref: "x", yref: "paper",
		x0: e.x, x1: e.x, y0: 0, y1: 1,
		line: { dash: "solid", width: Math.min(1 + Math.log2(e.size), 4), color: "rgba(220,0,0,0.5)" },
	})));

	const annotations = phases.map((p) => ({
		x: p.x, y: 1, yref: "paper",
		text: p.label, showarrow: false,
		font: { size: 10, color: "#555" },
	}));

	Plotly.newPlot("chart", traces, {
		barmode: "stack",
		margin: { t: 24, r: 16, l: 40, b: 32 },
		xaxis: { title: "сек" },
		yaxis: { title: "команд / 30 сек" },
		shapes,
		annotations,
		legend: { orientation: "h", y: -0.2 },
	}, { responsive: true });
})();
