package com.antclaw.alfq.ui.post

import com.antclaw.alfq.data.error.AppErrorCategory
import com.antclaw.alfq.ui.social.*
import com.antclaw.alfq.testutil.CoroutineTestBase
import com.antclaw.alfq.testutil.FakeSocialRepo
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.*
import org.junit.Assert.*

@OptIn(ExperimentalCoroutinesApi::class)
class PostViewModelTest : CoroutineTestBase() {

    @Test fun `post success transitions to Success state`() = runTest(scheduler) {
        val repo = FakeSocialRepo()
        val vm = PostViewModel(repo)
        vm.post(PostDraft(content = "hello"))
        advanceUntilIdle()
        assertTrue(vm.postState.value is PostState.Success)
    }

    @Test fun `post failure transitions to Error state with AppError`() = runTest(scheduler) {
        val repo = FakeSocialRepo(failCreate = true)
        val vm = PostViewModel(repo)
        vm.post(PostDraft(content = "hello"))
        advanceUntilIdle()
        val state = vm.postState.value
        assertTrue(state is PostState.Error)
        val err = (state as PostState.Error).error
        assertNotNull(err)
        assertEquals(AppErrorCategory.UNKNOWN, err.category)
    }

    @Test fun `retry re-submits last failed draft and succeeds`() = runTest(scheduler) {
        val repo = FakeSocialRepo(failCreate = true)
        val vm = PostViewModel(repo)
        val draft = PostDraft(content = "retry me", signalPair = "EURUSD", visibility = "followers")
        vm.post(draft)
        advanceUntilIdle()
        assertTrue(vm.postState.value is PostState.Error)
        repo.failCreate = false
        vm.retry()
        advanceUntilIdle()
        assertTrue(vm.postState.value is PostState.Success)
        assertEquals("retry me", repo.lastContent)
        assertEquals("EURUSD", repo.lastSignalPair)
        assertEquals("followers", repo.lastVisibility)
    }

    @Test fun `retry does nothing when no failed draft`() = runTest(scheduler) {
        val repo = FakeSocialRepo()
        val vm = PostViewModel(repo)
        vm.retry()
        advanceUntilIdle()
        assertEquals(0, repo.submitCount)
        assertTrue(vm.postState.value is PostState.Idle)
    }

    @Test fun `retry fails again keeps Error state`() = runTest(scheduler) {
        val repo = FakeSocialRepo(failCreate = true)
        val vm = PostViewModel(repo)
        vm.post(PostDraft(content = "fail1"))
        advanceUntilIdle()
        assertTrue(vm.postState.value is PostState.Error)
        vm.retry()
        advanceUntilIdle()
        assertTrue(vm.postState.value is PostState.Error)
    }

    @Test fun `loading guard prevents concurrent submit`() = runTest(scheduler) {
        val repo = FakeSocialRepo()
        val vm = PostViewModel(repo)
        vm.post(PostDraft(content = "first"))
        assertTrue(vm.postState.value is PostState.Loading)
        vm.post(PostDraft(content = "second"))
        assertTrue(vm.postState.value is PostState.Loading)
    }

    @Test fun `concurrent submit only calls createPost once`() = runTest(scheduler) {
        val repo = FakeSocialRepo()
        val vm = PostViewModel(repo)
        vm.post(PostDraft(content = "first"))
        vm.post(PostDraft(content = "second"))
        vm.post(PostDraft(content = "third"))
        advanceUntilIdle()
        assertEquals(1, repo.submitCount)
        assertEquals("first", repo.lastContent)
    }

    @Test fun `sequential posts each call createPost`() = runTest(scheduler) {
        val repo = FakeSocialRepo()
        val vm = PostViewModel(repo)
        vm.post(PostDraft(content = "one")); advanceUntilIdle(); vm.reset()
        vm.post(PostDraft(content = "two")); advanceUntilIdle(); vm.reset()
        vm.post(PostDraft(content = "three")); advanceUntilIdle()
        assertEquals(3, repo.submitCount)
    }

    @Test fun `reset clears state`() = runTest(scheduler) {
        val vm = PostViewModel(FakeSocialRepo())
        vm.post(PostDraft(content = "test"))
        assertTrue(vm.postState.value is PostState.Loading)
        vm.reset()
        assertTrue(vm.postState.value is PostState.Idle)
    }

    @Test fun `reset after success clears state`() = runTest(scheduler) {
        val repo = FakeSocialRepo()
        val vm = PostViewModel(repo)
        vm.post(PostDraft(content = "ok"))
        advanceUntilIdle()
        assertTrue(vm.postState.value is PostState.Success)
        vm.reset()
        assertTrue(vm.postState.value is PostState.Idle)
    }
}

