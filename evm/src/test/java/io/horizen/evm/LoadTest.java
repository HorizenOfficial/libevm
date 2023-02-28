package io.horizen.evm;

import org.junit.Ignore;
import org.junit.Test;

import java.util.TimerTask;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

import static org.junit.Assert.assertTrue;

public class LoadTest extends LibEvmTestBase {

    private static final AtomicInteger counter = new AtomicInteger(0);
    private static final long BYTE_TO_MB_CONVERSION_VALUE = 1024 * 1024;
    private static final EvmTest test = new EvmTest();

    private static long getCurrentlyUsedMemory() {
        System.gc();
        return (Runtime.getRuntime().totalMemory() - Runtime.getRuntime().freeMemory()) / BYTE_TO_MB_CONVERSION_VALUE;
    }

    private static class StatusTask extends TimerTask {
        @Override
        public void run() {
            System.err.printf("used memory: %d MB iterations: %d%n", getCurrentlyUsedMemory(), counter.get());
        }
    }

    private static class LoadTask extends TimerTask {
        @Override
        public void run() {
            try {
                test.blockHashCallback();
                counter.getAndIncrement();
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        }
    }

    private static ExecutorService startLoadTasks(int tasks) {
        var executor = Executors.newScheduledThreadPool(tasks);
        for (int i = 0; i < tasks; i++) {
            executor.scheduleAtFixedRate(new LoadTask(), 1000, 100, TimeUnit.MILLISECONDS);
        }
        executor.scheduleAtFixedRate(new StatusTask(), 0, 1000, TimeUnit.MILLISECONDS);
        return executor;
    }

    private static void runLoadTest(int tasks, int duration) throws Exception {
        var executor = startLoadTasks(tasks);
        Thread.sleep(duration);
        executor.shutdown();
        assertTrue("timeout while stopping load tasks", executor.awaitTermination(1000, TimeUnit.MILLISECONDS));
    }

    @Test
    @Ignore
    public void lowLoad() throws Exception {
        runLoadTest(1, 3000);
    }

    @Test
    @Ignore
    public void highLoad() throws Exception {
        runLoadTest(32, 10000);
    }

    public static void main(String[] args) throws Exception {
        var tasks = 32;
        if (args.length > 0) {
            tasks = Integer.parseInt(args[0]);
        }
        System.err.printf("starting load tasks: %d%n", tasks);
        var executor = startLoadTasks(tasks);
        try {
            // wait until enter key
            System.in.read();
        } finally {
            System.err.println("interrupted, stopping load tasks");
            executor.shutdown();
            if (!executor.awaitTermination(1000, TimeUnit.MILLISECONDS)) {
                System.err.println("timeout while stopping load tasks");
            }
        }
    }
}
