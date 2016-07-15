
import java.io.PrintStream

import org.HdrHistogram.{Recorder,HistogramLogWriter}

object Main {
	def main(args: Array[String]): Unit = {
		val testCases = Array(
			("single", Array(100, 1000, 235, 10000)),
			("single_repeated", Array(100, 100, 100, 500, 3)),
			("single_repeated_multi", Array(100, 100, 100, 200, 200, 423, 512, 100000000, 1200000, 2)))

		testCases.map(
			{ case (name, vals) => {
				val r = new Recorder(3)
				vals.map(v => r.recordValue(v))
				val h = r.getIntervalHistogram()
				val w = new HistogramLogWriter("../" + name + ".log")
				w.outputIntervalHistogram(h)
				val ans = new PrintStream("../" + name + ".ans")
				ans.println(vals.length)
				val counts = vals
					.map(v => (v, 1))
					.groupBy(_._1)
					.map({ case (v: Int, cs: Array[(Int, Int)]) => (v, cs.map(_._2).reduce(_+_)) })
				counts.map({ case (v, c) => ans.println("%d,%d".format(v, c)) })
			}}
		)
		/*
		val r = new Recorder(3)
		r.recordValue(100)
		r.recordValue(1000)
		r.recordValue(235)
		r.recordValue(10000)

		val h = r.getIntervalHistogram()
		val w = new HistogramLogWriter("../single_hist.log")
		w.outputIntervalHistogram(h)
		val golden = new PrintStream("../single_hist.ans")
		golden.println(4)
		golden.println("%d,%d".format(100, 1))
		golden.println("%d,%d".format(1000, 1))
		golden.println("%d,%d".format(235, 1))
		golden.println("%d,%d".format(10000, 1))
		*/
	}
}