import java.io.PrintStream

import org.HdrHistogram.{Recorder, HistogramLogWriter}

object Main {
  def main(args: Array[String]): Unit = {
    genReader()
    genWriter()
    genTimestamp()
  }

  def genReader(): Unit = {
    val testCases = List(
        ("single", List(100, 1000, 235, 10000)),
        ("single_repeated", List(100, 100, 100, 500, 3)),
        ("single_repeated_multi",
         List(100, 100, 100, 200, 200, 423, 512, 100000000, 1200000, 2)))

    testCases.map({
      case (name, vals) => {
        val r = new Recorder(3)
        vals.map(v => r.recordValue(v))
        val h = r.getIntervalHistogram()
        val w = new HistogramLogWriter("../" + name + ".log")
        h.setStartTimeStamp(100)
        h.setEndTimeStamp(3000)
        w.outputIntervalHistogram(h)
        val ans = new PrintStream("../" + name + ".ans")
        ans.println(vals.length)
        val counts = vals
          .map(v => (v, 1))
          .groupBy(_._1)
          .map({
            case (v: Int, cs: List[(Int, Int)]) =>
              (v, cs.map(_._2).reduce(_ + _))
          })
        counts.map({ case (v, c) => ans.println("%d,%d".format(v, c)) })
        ans.close()
      }
    })
  }

  def genWriter(): Unit = {
    val testCases = List(
        ("wsingle", List(List(123, 1455, 232, 90999))),
        ("wsingle_repeated", List(List(123, 123, 123, 641, 2))),
        ("wsingle_repeated_multi",
         List(List(123, 123, 145, 145, 555, 666, 666, 1400000, 540000, 9, 1))),
        ("wmulti",
         List(List(555, 21, 8, 132123, 11),
              List(789, 9, 1),
              List(),
              List(0),
              List(1))),
        ("wmulti_repeated",
         List(List(12, 11, 12), List(1, 1), List(1, 1), List())))

    testCases.map({
      case (name, valslist) => {
        val r = new Recorder(3)
        val w = new HistogramLogWriter("../" + name + ".golden")
        val ins = new PrintStream("../" + name + ".ins")
        valslist.zipWithIndex.map({
          case (vals, index) => {
            ins.println(index) // start time
            ins.println(index + 1) // end time
            vals.map(r.recordValue(_))
            vals.map(v => ins.println(v))
            ins.println("---")
            val h = r.getIntervalHistogram()
            h.setStartTimeStamp(index)
            h.setEndTimeStamp(index + 1)
            w.outputIntervalHistogram(h)
          }
        })
        ins.close()
      }
    })
  }

  def genTimestamp(): Unit = {
    val StartTime = 123L
    val BaseTime = 14124L

    val HistStart = 15000L
    val HistEnd = 16003L

    val r = new Recorder(3);
    val w = new HistogramLogWriter("../tstamp.log")
    List(100, 10, 44444).map(r.recordValue(_))

    val h = r.getIntervalHistogram()
    h.setStartTimeStamp(HistStart)
    h.setEndTimeStamp(HistEnd)

    w.outputStartTime(StartTime)
    w.setBaseTime(BaseTime)
    w.outputBaseTime(BaseTime)
    w.outputComment("this should be ignored")
    w.outputLogFormatVersion()
    w.outputLegend()
    w.outputIntervalHistogram(h)
  }
}

