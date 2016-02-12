var starts = _.sortBy(allData.Start, function(d) {
  return d.CompletionTime
});

var skippables = _.map(starts.slice(0, -1), function(d, i) {
  return (starts[i + 1].CompletionTime - d.CompletionTime) > 20 ? i : 0;
});
var skippablesIdx = _.filter(skippables, function(d) {
  return d > 0;
});


// generates title

var events = _.sortBy(_.map(allData.EventLog,
  function(ev) {
    return {
      cmd: ev.Cmd + " " + ((ev.Args != null) ? ev.Args.join(" ") : ""),
      ts: ev.StartTime,
      end: ev.EndTime
    };
  }), function(x) {return x.ts});



var title = d3.select("#content").append("h4")

var stringify = function(d) {
  return _.flatten(_.map(_.keys(d), function(k) {
    return [k, ": ", d[k], "<br/>"]
  })).join("");
}

var broken = [];

var lastBreak = 0
_.each(starts.slice(0, -1),
  function(d, i) {
    if ((starts[i + 1].CompletionTime - d.CompletionTime) > 20) {
      broken.push(starts.slice(lastBreak, i));
      lastBreak = i
    }
  });
broken.push(starts.slice(lastBreak, -1));

var margin = {
    top: 20,
    right: 50,
    bottom: 30,
    left: 30
  },
  width = 960 - margin.left - margin.right,
  height = 500 - margin.top - margin.bottom;

var xScale = d3.scale.linear()
  .domain([0, 1.01 * d3.max(allData.EventLog, function(d) {
    return d.EndTime;
  })])
  .range([0, width]);

var yScale = d3.scale.linear()
  .domain([0, 1.01 * d3.max(starts, function(d) {
    return d.Delay;
  })])
  .range([height, 0]);

var yUnitScale = d3.scale.linear()
  .domain([0, d3.max(starts, function(d) {
    return d.RunningCount;
  })])
  .range([height, 0]);

var yCPUScale = d3.scale.linear()
  .domain([0, 100])
  .range([height, 0]);

var xAxis = d3.svg.axis()
  .scale(xScale)
  .orient("bottom")
  .innerTickSize(-height)
  .outerTickSize(0)
  .tickPadding(10);

var yAxis = d3.svg.axis()
  .scale(yScale)
  .ticks(10)
  .orient("left")
  .innerTickSize(-width)
  .outerTickSize(0);

var yUnitsAxis = d3.svg.axis()
  .scale(yUnitScale)
  .ticks(10)
  .orient("right")
  .innerTickSize(-width)
  .outerTickSize(0);

var canvas = d3.select("#content")
  .append("svg")
  .attr("width", width + margin.left + margin.right)
  .attr("height", height + margin.top + margin.bottom)
  .append("g")
  .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

canvas.append("g")
  .attr("class", "x axis")
  .attr("transform", "translate(0," + height + ")")
  .call(xAxis);

canvas.append("g")
  .attr("class", "y axis left")
  .call(yAxis);

canvas.append("g")
  .attr("class", "y axis nogrid")
  .attr("transform", "translate(" + width + ",0)")
  .call(yUnitsAxis);


var div = d3.select("body").append("div")
  .attr("class", "tooltip")
  .style("opacity", 0);

var bars = canvas.selectAll("circles")
  .data(allData.Start)
  .enter()
  .append("circle")
  .attr("class", "delaypoint")
  .attr("cy", function(d) {
    return yScale(d.Delay);
  })
  .attr("r", function(d) {
    return 2;
  })
  .attr("cx", function(d, i) {
    return xScale(d.CompletionTime);
  })
  .on("mouseover", function(d) {
    div.transition()
      .duration(20)
      .style("opacity", .9);
    div.html(stringify(d))
      .style("left", (d3.event.pageX) + "px")
      .style("top", (d3.event.pageY - 28) + "px");
  })
  .on("mouseout", function(d) {
    div.transition()
      .duration(500)
      .style("opacity", 0);
  });

_.each(broken, function(data) {
  var startingCount = canvas.append("path")
    .data([data.slice(1, -1)])
    .attr("class", "line-starting-count")
    .attr("d", d3.svg.line()
      .x(function(d) {
        return xScale(d.CompletionTime);
      })
      .y(function(d) {
        return yUnitScale(d.StartingCount);
      })
    );
})

var runningCount = canvas.append("path")
  .data([starts])
  .attr("class", "line-running-count")
  .attr("d", d3.svg.line()
    .x(function(d) {
      return xScale(d.CompletionTime);
    })
    .y(function(d) {
      return yUnitScale(d.RunningCount);
    })
  );

_.each(allData.MachineStats, function(machineStats, machineName) {
  var systemdLine = _.filter(machineStats, function(obj){ return obj.Process == "systemd";});
  canvas.append("path")
    .data([systemdLine])
    .attr("class", "systemd-cpu-usage")
    .attr("d", d3.svg.line()
        .x(function(d) {
          return xScale(d.TimeStamp);
        })
        .y(function(d) {
          return yCPUScale(d.CPUUsage);
        })
        );
  var fleetLine = _.filter(machineStats, function(obj){ return obj.Process == "fleetd";});
  canvas.append("path")
    .data([fleetLine])
    .attr("class", "fleet-cpu-usage")
    .attr("d", d3.svg.line()
        .x(function(d) {
          return xScale(d.TimeStamp);
        })
        .y(function(d) {
          return yCPUScale(d.CPUUsage);
        })
        );
})


// labels

var xlabel = canvas.append("text")
  .attr("transform", "translate(" + (width / 2) + " ," + (height + margin.bottom) + ")")
  .style("text-anchor", "middle")
  .text("test time (s)");

var ylabel = canvas.append("text")
  .attr("transform", "rotate(-90)")
  .attr("y", 0 - margin.left)
  .attr("x", 0 - (height / 2))
  .attr("dy", "1em")
  .style("text-anchor", "middle")
  .text("delay (s)");

var yUnitslabel = canvas.append("text")
  .attr("transform", "rotate(-90)")
  .attr("y", width + margin.right / 2)
  .attr("x", 0 - (height / 2))
  .attr("dy", "1em")
  .style("text-anchor", "middle")
  .text("unit count");

var createLegend = function(g) {
  var legend = g.append("g")
    .attr("class", "legend")

  legend.append("circle").attr("class", "delaypoint")
    .attr("cx", width / 2).attr("cy", 0).attr("r", 2)

  var createLabel = function(label, y_offset) {
    return function(g) {
      g.append("text")
        .attr("x", width / 2 + 5)
        .attr("y", y_offset)
        .attr("dy", ".35em")
        .style("text-anchor", "begin")
        .text(label)
    };
  };

  var createLegendLine = function(klass, y_offset) {
    return function(g) {
      g.append("line")
        .attr("class", klass)
        .attr("x1", width / 2 - 5)
        .attr("x2", width / 2 + 3)
        .attr("y1", y_offset)
        .attr("y2", y_offset)
    };
  }

  legend.call(createLabel("delay between star-trigger and real-start", 0));
  legend.call(createLabel("units running", 15));
  legend.call(createLabel("units starting", 30));
  legend.call(createLegendLine("line-running-count", 15))
  legend.call(createLegendLine("line-starting-count", 30))
};

canvas.call(createLegend)

var eventLine = canvas.append("rect")
  .attr("class","timeline-focus")
  .attr("x", 0)
  .attr("y", 0)
  .attr("width", 0)
  .attr("height", height)
  .style("display","none");

_.each(events, function(ev) {
  title
    .append("span")
    .attr("class","event-text")
    .on("mouseover", function() { eventLine.style("display", null); })
    .on("mouseout", function() { eventLine.style("display", "none"); })
    .on("mousemove",function() {
      eventLine.attr("x",xScale(ev.ts))
      eventLine.attr("width", xScale(ev.end) - xScale(ev.ts))
    })
    .text(ev.cmd);
});
