import React, { useCallback, useMemo, useRef, useState } from 'react';
import { CartesianCoords2D, DataFrame, FieldType, PanelProps } from '@grafana/data';
import { Portal, UPlotConfigBuilder, useTheme2, VizTooltipContainer, ZoomPlugin } from '@grafana/ui';
import { TimelineMode, TimelineOptions } from './types';
import { TimelineChart } from './TimelineChart';
import { prepareTimelineFields, prepareTimelineLegendItems } from './utils';
import { StateTimelineTooltip } from './StateTimelineTooltip';
import { getLastStreamingDataFramePacket } from '@grafana/data/src/dataframe/StreamingDataFrame';
import { HoverEvent, setupConfig } from '../barchart/config';

const TOOLTIP_OFFSET = 10;

interface TimelinePanelProps extends PanelProps<TimelineOptions> {}

/**
 * @alpha
 */
export const StateTimelinePanel: React.FC<TimelinePanelProps> = ({
  data,
  timeRange,
  timeZone,
  options,
  width,
  height,
  onChangeTimeRange,
}) => {
  const theme = useTheme2();

  const oldConfig = useRef<UPlotConfigBuilder | undefined>(undefined);
  const isToolTipOpen = useRef<boolean>(false);

  const [hover, setHover] = useState<HoverEvent | undefined>(undefined);
  const [coords, setCoords] = useState<CartesianCoords2D | null>(null);
  const [focusedSeriesIdx, setFocusedSeriesIdx] = useState<number | null>(null);
  const [focusedPointIdx, setFocusedPointIdx] = useState<number | null>(null);
  const [_dummy, setDummyToForceRender] = useState<boolean>(false);

  const onUPlotClick = () => {
    isToolTipOpen.current = !isToolTipOpen.current;

    // Linking into useState required to re-render tooltip
    setDummyToForceRender(isToolTipOpen.current);
  };

  const { frames, warn } = useMemo(() => prepareTimelineFields(data?.series, options.mergeValues ?? true, theme), [
    data,
    options.mergeValues,
    theme,
  ]);

  const legendItems = useMemo(() => prepareTimelineLegendItems(frames, options.legend, theme), [
    frames,
    options.legend,
    theme,
  ]);

  const renderCustomTooltip = useCallback(
    (alignedData: DataFrame, seriesIdx: number | null, datapointIdx: number | null) => {
      const data = frames ?? [];
      // Count value fields in the state-timeline-ready frame
      const valueFieldsCount = data.reduce(
        (acc, frame) => acc + frame.fields.filter((field) => field.type !== FieldType.time).length,
        0
      );

      // Not caring about multi mode in StateTimeline
      if (seriesIdx === null || datapointIdx === null) {
        return null;
      }

      /**
       * There could be a case when the tooltip shows a data from one of a multiple query and the other query finishes first
       * from refreshing. This causes data to be out of sync. alignedData - 1 because Time field doesn't count.
       * Render nothing in this case to prevent error.
       * See https://github.com/grafana/support-escalations/issues/932
       */
      if (
        (!alignedData.meta?.transformations?.length && alignedData.fields.length - 1 !== valueFieldsCount) ||
        !alignedData.fields[seriesIdx]
      ) {
        return null;
      }

      return (
        <StateTimelineTooltip
          data={data}
          alignedData={alignedData}
          seriesIdx={seriesIdx}
          datapointIdx={datapointIdx}
          timeZone={timeZone}
        />
      );
    },
    [timeZone, frames]
  );

  if (!frames || warn) {
    return (
      <div className="panel-empty">
        <p>{warn ?? 'No data found in response'}</p>
      </div>
    );
  }

  if (frames.length === 1) {
    const packet = getLastStreamingDataFramePacket(frames[0]);
    if (packet) {
      // console.log('STREAM Packet', packet);
    }
  }

  return (
    <TimelineChart
      theme={theme}
      frames={frames}
      structureRev={data.structureRev}
      timeRange={timeRange}
      timeZone={timeZone}
      width={width}
      height={height}
      legendItems={legendItems}
      {...options}
      mode={TimelineMode.Changes}
    >
      {(config, alignedFrame) => {
        if (oldConfig.current !== config) {
          oldConfig.current = setupConfig({
            config,
            onUPlotClick,
            setFocusedSeriesIdx,
            setFocusedPointIdx,
            setCoords,
            setHover,
            isToolTipOpen,
          });
        }
        return (
          <>
            <ZoomPlugin config={config} onZoom={onChangeTimeRange} />
            <Portal>
              {hover && coords && (
                <VizTooltipContainer
                  position={{ x: coords.x, y: coords.y }}
                  offset={{ x: TOOLTIP_OFFSET, y: TOOLTIP_OFFSET }}
                  allowPointerEvents={isToolTipOpen.current}
                >
                  {renderCustomTooltip(alignedFrame, focusedSeriesIdx, focusedPointIdx)}
                </VizTooltipContainer>
              )}
            </Portal>
            <OutsideRangePlugin config={config} onChangeTimeRange={onChangeTimeRange} />
          </>
        );
      }}
    </TimelineChart>
  );
};
